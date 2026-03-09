package uritemplate

import (
	"fmt"
	"io"
	"strings"
)

// encType controls percent-encoding behavior.
type encType int

const (
	encU  encType = iota // encode all except unreserved
	encUR                // encode all except unreserved + reserved
)

// modifier represents a variable modifier type.
type modifier int

const (
	modNone    modifier = iota
	modPrefix           // :N prefix modifier
	modExplode          // * explode modifier
)

// operator represents the expression operator (+, #, ., /, ;, ?, &, or none).
type operator struct {
	prefix string  // first character prefix ("", "#", ".", "/", ";", "?", "&")
	sep    string  // separator between expanded variables
	named  bool    // whether to include var=value naming
	ifemp  string  // value to use when variable is empty (for named)
	allow  encType // encoding: unreserved only (U) or unreserved+reserved (U+R)
}

var operators = map[byte]operator{
	0:   {prefix: "", sep: ",", named: false, ifemp: "", allow: encU},
	'+': {prefix: "", sep: ",", named: false, ifemp: "", allow: encUR},
	'#': {prefix: "#", sep: ",", named: false, ifemp: "", allow: encUR},
	'.': {prefix: ".", sep: ".", named: false, ifemp: "", allow: encU},
	'/': {prefix: "/", sep: "/", named: false, ifemp: "", allow: encU},
	';': {prefix: ";", sep: ";", named: true, ifemp: "", allow: encU},
	'?': {prefix: "?", sep: "&", named: true, ifemp: "=", allow: encU},
	'&': {prefix: "&", sep: "&", named: true, ifemp: "=", allow: encU},
}

// varSpec represents a variable with optional modifier.
type varSpec struct {
	name     string
	modifier modifier
	maxLen   int // only used with modPrefix
}

// part is either a literal string or an expression.
type part interface {
	expand(w io.Writer, vars Values) error
}

// literal is a plain string part of a template.
type literal string

// expression is a template expression like {var}, {+var}, {#var}, etc.
type expression struct {
	operator operator
	vars     []varSpec
}

// Template is a parsed URI template consisting of literal and expression parts.
type Template struct {
	raw   string
	parts []part
}

// String returns the original template string.
func (t *Template) String() string {
	return t.raw
}

// Parse parses a URI Template string per RFC 6570.
// Returns an error if the template is malformed.
func Parse(template string) (*Template, error) {
	t := &Template{raw: template}
	i := 0
	n := len(template)
	var litBuf strings.Builder

	for i < n {
		ch := template[i]
		if ch == '}' {
			return nil, fmt.Errorf("uritemplate: unexpected '}' at position %d", i)
		}
		if ch == '{' {
			// flush literal
			if litBuf.Len() > 0 {
				t.parts = append(t.parts, literal(litBuf.String()))
				litBuf.Reset()
			}
			// find closing brace
			end := strings.IndexByte(template[i:], '}')
			if end < 0 {
				return nil, fmt.Errorf("uritemplate: unclosed '{' at position %d", i)
			}
			end += i // absolute position
			body := template[i+1 : end]
			if len(body) == 0 {
				return nil, fmt.Errorf("uritemplate: empty expression at position %d", i)
			}
			expr, err := parseExpression(body, i)
			if err != nil {
				return nil, err
			}
			t.parts = append(t.parts, expr)
			i = end + 1
			continue
		}
		litBuf.WriteByte(ch)
		i++
	}

	if litBuf.Len() > 0 {
		t.parts = append(t.parts, literal(litBuf.String()))
	}

	return t, nil
}

func parseExpression(body string, pos int) (*expression, error) {
	expr := &expression{}

	// Check for operator
	offset := 0
	if len(body) > 0 {
		first := body[0]
		if op, ok := operators[first]; ok {
			expr.operator = op
			offset = 1
		} else if first == '+' || first == '#' || first == '.' || first == '/' || first == ';' || first == '?' || first == '&' {
			// already handled above via map lookup
			// This branch won't be reached for valid operators
			return nil, fmt.Errorf("uritemplate: invalid operator '%c' at position %d", first, pos)
		} else if !isVarStart(first) && first != '%' {
			return nil, fmt.Errorf("uritemplate: invalid operator '%c' at position %d", first, pos+1)
		} else {
			expr.operator = operators[0]
		}
	}

	varList := body[offset:]
	if len(varList) == 0 {
		return nil, fmt.Errorf("uritemplate: empty expression at position %d", pos)
	}

	specs := strings.Split(varList, ",")
	for _, spec := range specs {
		vs, err := parseVarSpec(spec, pos)
		if err != nil {
			return nil, err
		}
		expr.vars = append(expr.vars, vs)
	}

	return expr, nil
}

func isVarStart(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' || (c >= '0' && c <= '9')
}

func isVarChar(c byte) bool {
	return isVarStart(c) || c == '.'
}

func parseVarSpec(spec string, pos int) (varSpec, error) {
	if len(spec) == 0 {
		return varSpec{}, fmt.Errorf("uritemplate: empty variable name at position %d", pos)
	}

	vs := varSpec{}

	// Check for explode modifier
	if strings.HasSuffix(spec, "*") {
		vs.modifier = modExplode
		spec = spec[:len(spec)-1]
	}

	// Check for prefix modifier
	if colonIdx := strings.IndexByte(spec, ':'); colonIdx >= 0 {
		if vs.modifier == modExplode {
			return varSpec{}, fmt.Errorf("uritemplate: variable cannot have both prefix and explode modifiers at position %d", pos)
		}
		prefixStr := spec[colonIdx+1:]
		spec = spec[:colonIdx]
		vs.modifier = modPrefix

		if len(prefixStr) == 0 {
			return varSpec{}, fmt.Errorf("uritemplate: empty prefix length at position %d", pos)
		}
		maxLen := 0
		for _, c := range prefixStr {
			if c < '0' || c > '9' {
				return varSpec{}, fmt.Errorf("uritemplate: invalid prefix length at position %d", pos)
			}
			maxLen = maxLen*10 + int(c-'0')
		}
		if maxLen < 1 || maxLen > 9999 {
			return varSpec{}, fmt.Errorf("uritemplate: prefix length must be 1-9999 at position %d", pos)
		}
		vs.maxLen = maxLen
	}

	// Validate variable name
	if err := validateVarName(spec, pos); err != nil {
		return varSpec{}, err
	}
	vs.name = spec

	return vs, nil
}

func validateVarName(name string, pos int) error {
	if len(name) == 0 {
		return fmt.Errorf("uritemplate: empty variable name at position %d", pos)
	}

	// Variable name must not start with '.'
	if name[0] == '.' {
		return fmt.Errorf("uritemplate: variable name cannot start with '.' at position %d", pos)
	}

	i := 0
	for i < len(name) {
		c := name[i]
		if c == '%' {
			// pct-encoded: must be %HH
			if i+2 >= len(name) || !isHex(name[i+1]) || !isHex(name[i+2]) {
				return fmt.Errorf("uritemplate: invalid percent-encoding in variable name at position %d", pos)
			}
			i += 3
			continue
		}
		if c == '.' {
			// dot must not be last character and must be followed by varchar
			if i+1 >= len(name) {
				return fmt.Errorf("uritemplate: variable name cannot end with '.' at position %d", pos)
			}
			next := name[i+1]
			if !isVarStart(next) && next != '%' {
				return fmt.Errorf("uritemplate: invalid character after '.' in variable name at position %d", pos)
			}
			i++
			continue
		}
		if !isVarStart(c) {
			return fmt.Errorf("uritemplate: invalid character '%c' in variable name at position %d", c, pos)
		}
		i++
	}

	return nil
}

func isHex(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}
