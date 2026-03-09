package uritemplate

import (
	"strings"
)

// Match attempts to extract variable values from a URI string by matching it
// against this template. Returns the extracted values and true if the URI
// matches the template pattern, or nil and false otherwise.
//
// Matching works reliably for templates where expressions are delimited by
// literal characters, the start/end of the string, or operator-specific
// delimiters. Ambiguous templates (e.g., two adjacent simple string expressions
// without separating literals) may not produce correct results.
func (t *Template) Match(uri string) (Values, bool) {
	m := &matcher{
		template: t,
		values:   make(Values),
	}
	if m.match(uri) {
		return m.values, true
	}
	return nil, false
}

type matcher struct {
	template *Template
	values   Values
}

func (m *matcher) match(uri string) bool {
	pos := 0
	for i, p := range m.template.parts {
		switch part := p.(type) {
		case literal:
			lit := string(part)
			if !strings.HasPrefix(uri[pos:], lit) {
				return false
			}
			pos += len(lit)
		case *expression:
			boundary := m.findBoundary(uri, pos, i)
			if boundary < pos {
				return false
			}
			segment := uri[pos:boundary]
			if !m.matchExpression(part, segment) {
				return false
			}
			pos = boundary
		}
	}
	return pos == len(uri)
}

func (m *matcher) findBoundary(uri string, pos int, partIndex int) int {
	allowReserved := false
	if expr, ok := m.template.parts[partIndex].(*expression); ok {
		allowReserved = expr.operator.allow == encUR
	}

	remaining := uri[pos:]
	for j := partIndex + 1; j < len(m.template.parts); j++ {
		switch next := m.template.parts[j].(type) {
		case literal:
			var idx int
			if allowReserved {
				idx = strings.LastIndex(remaining, string(next))
			} else {
				idx = strings.Index(remaining, string(next))
			}
			if idx < 0 {
				return -1
			}
			return pos + idx
		case *expression:
			if next.operator.prefix != "" {
				var idx int
				if allowReserved {
					idx = strings.LastIndex(remaining, next.operator.prefix)
				} else {
					idx = strings.Index(remaining, next.operator.prefix)
				}
				if idx >= 0 {
					return pos + idx
				}
				// prefix not found — expression may have expanded to nothing
				continue
			}
			continue
		}
	}
	return len(uri)
}

func (m *matcher) matchExpression(expr *expression, segment string) bool {
	op := expr.operator

	if segment == "" {
		return true
	}

	if op.prefix != "" {
		if !strings.HasPrefix(segment, op.prefix) {
			return false
		}
		segment = segment[len(op.prefix):]
	}

	if op.named {
		return m.matchNamed(expr, segment, op)
	}
	return m.matchPositional(expr, segment, op)
}

func (m *matcher) matchNamed(expr *expression, segment string, op operator) bool {
	parts := strings.Split(segment, op.sep)

	varSpecs := make(map[string]*varSpec)
	for i := range expr.vars {
		varSpecs[expr.vars[i].name] = &expr.vars[i]
	}

	explodeLists := make(map[string][]string)
	explodeKeys := make(map[string][]KeyValue)

	for _, part := range parts {
		if part == "" {
			continue
		}
		eqIdx := strings.IndexByte(part, '=')
		var name, value string
		if eqIdx >= 0 {
			name = pctDecode(part[:eqIdx])
			value = pctDecode(part[eqIdx+1:])
		} else {
			name = pctDecode(part)
			value = ""
		}

		if vs, ok := varSpecs[name]; ok {
			if vs.modifier == modExplode {
				explodeLists[name] = append(explodeLists[name], value)
			} else {
				m.values[name] = String(value)
			}
		} else {
			// Name doesn't match any varSpec — might be exploded keys
			for _, vs := range expr.vars {
				if vs.modifier == modExplode {
					explodeKeys[vs.name] = append(explodeKeys[vs.name], KeyValue{name, value})
					break
				}
			}
		}
	}

	for name, values := range explodeLists {
		if len(values) == 1 {
			m.values[name] = String(values[0])
		} else {
			m.values[name] = List(values...)
		}
	}

	for name, pairs := range explodeKeys {
		m.values[name] = Keys(pairs...)
	}

	return true
}

func (m *matcher) matchPositional(expr *expression, segment string, op operator) bool {
	if len(expr.vars) == 1 && expr.vars[0].modifier == modExplode {
		vs := expr.vars[0]
		parts := strings.Split(segment, op.sep)

		hasEquals := true
		for _, p := range parts {
			if !strings.Contains(p, "=") {
				hasEquals = false
				break
			}
		}

		if hasEquals && len(parts) > 0 {
			var pairs []KeyValue
			for _, p := range parts {
				eqIdx := strings.IndexByte(p, '=')
				k := pctDecode(p[:eqIdx])
				v := pctDecode(p[eqIdx+1:])
				pairs = append(pairs, KeyValue{k, v})
			}
			m.values[vs.name] = Keys(pairs...)
		} else {
			decoded := make([]string, len(parts))
			for i, p := range parts {
				decoded[i] = pctDecode(p)
			}
			m.values[vs.name] = List(decoded...)
		}
		return true
	}

	parts := strings.Split(segment, op.sep)

	for i, vs := range expr.vars {
		if i >= len(parts) {
			break
		}
		m.values[vs.name] = String(pctDecode(parts[i]))
	}

	return true
}

func pctDecode(s string) string {
	if !strings.Contains(s, "%") {
		return s
	}

	var buf strings.Builder
	buf.Grow(len(s))
	for i := 0; i < len(s); {
		if s[i] == '%' && i+2 < len(s) && isHex(s[i+1]) && isHex(s[i+2]) {
			hi := unhex(s[i+1])
			lo := unhex(s[i+2])
			buf.WriteByte(hi<<4 | lo)
			i += 3
		} else {
			buf.WriteByte(s[i])
			i++
		}
	}
	return buf.String()
}

func unhex(c byte) byte {
	switch {
	case c >= '0' && c <= '9':
		return c - '0'
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10
	}
	return 0
}
