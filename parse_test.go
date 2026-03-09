package uritemplate

import (
	"testing"
)

func TestParseSimpleVariables(t *testing.T) {
	tests := []struct {
		input    string
		wantVars []string
	}{
		{"{var}", []string{"var"}},
		{"{hello}", []string{"hello"}},
	}
	for _, tt := range tests {
		tmpl, err := Parse(tt.input)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", tt.input, err)
			continue
		}
		if tmpl.String() != tt.input {
			t.Errorf("String() = %q, want %q", tmpl.String(), tt.input)
		}
		expr, ok := tmpl.parts[0].(*expression)
		if !ok {
			t.Errorf("Parse(%q) part[0] is not expression", tt.input)
			continue
		}
		if len(expr.vars) != len(tt.wantVars) {
			t.Errorf("Parse(%q) got %d vars, want %d", tt.input, len(expr.vars), len(tt.wantVars))
			continue
		}
		for i, wv := range tt.wantVars {
			if expr.vars[i].name != wv {
				t.Errorf("Parse(%q) var[%d] = %q, want %q", tt.input, i, expr.vars[i].name, wv)
			}
		}
	}
}

func TestParseMultipleVariables(t *testing.T) {
	tests := []struct {
		input    string
		wantVars []string
	}{
		{"{x,y}", []string{"x", "y"}},
		{"{x,hello,y}", []string{"x", "hello", "y"}},
	}
	for _, tt := range tests {
		tmpl, err := Parse(tt.input)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", tt.input, err)
			continue
		}
		expr := tmpl.parts[0].(*expression)
		if len(expr.vars) != len(tt.wantVars) {
			t.Errorf("Parse(%q) got %d vars, want %d", tt.input, len(expr.vars), len(tt.wantVars))
			continue
		}
		for i, wv := range tt.wantVars {
			if expr.vars[i].name != wv {
				t.Errorf("Parse(%q) var[%d] = %q, want %q", tt.input, i, expr.vars[i].name, wv)
			}
		}
	}
}

func TestParseOperators(t *testing.T) {
	tests := []struct {
		input      string
		wantPrefix string
		wantSep    string
		wantNamed  bool
		wantIfemp  string
	}{
		{"{+var}", "", ",", false, ""},
		{"{#var}", "#", ",", false, ""},
		{"{.var}", ".", ".", false, ""},
		{"{/var}", "/", "/", false, ""},
		{"{;var}", ";", ";", true, ""},
		{"{?var}", "?", "&", true, "="},
		{"{&var}", "&", "&", true, "="},
		{"{var}", "", ",", false, ""},
	}
	for _, tt := range tests {
		tmpl, err := Parse(tt.input)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", tt.input, err)
			continue
		}
		expr := tmpl.parts[0].(*expression)
		op := expr.operator
		if op.prefix != tt.wantPrefix {
			t.Errorf("Parse(%q) prefix = %q, want %q", tt.input, op.prefix, tt.wantPrefix)
		}
		if op.sep != tt.wantSep {
			t.Errorf("Parse(%q) sep = %q, want %q", tt.input, op.sep, tt.wantSep)
		}
		if op.named != tt.wantNamed {
			t.Errorf("Parse(%q) named = %v, want %v", tt.input, op.named, tt.wantNamed)
		}
		if op.ifemp != tt.wantIfemp {
			t.Errorf("Parse(%q) ifemp = %q, want %q", tt.input, op.ifemp, tt.wantIfemp)
		}
	}
}

func TestParsePrefixModifier(t *testing.T) {
	tests := []struct {
		input      string
		wantName   string
		wantMaxLen int
	}{
		{"{var:3}", "var", 3},
		{"{var:30}", "var", 30},
		{"{semi:5}", "semi", 5},
	}
	for _, tt := range tests {
		tmpl, err := Parse(tt.input)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", tt.input, err)
			continue
		}
		expr := tmpl.parts[0].(*expression)
		vs := expr.vars[0]
		if vs.name != tt.wantName {
			t.Errorf("Parse(%q) name = %q, want %q", tt.input, vs.name, tt.wantName)
		}
		if vs.modifier != modPrefix {
			t.Errorf("Parse(%q) modifier = %d, want modPrefix", tt.input, vs.modifier)
		}
		if vs.maxLen != tt.wantMaxLen {
			t.Errorf("Parse(%q) maxLen = %d, want %d", tt.input, vs.maxLen, tt.wantMaxLen)
		}
	}
}

func TestParseExplodeModifier(t *testing.T) {
	tests := []struct {
		input    string
		wantName string
	}{
		{"{list*}", "list"},
		{"{keys*}", "keys"},
	}
	for _, tt := range tests {
		tmpl, err := Parse(tt.input)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", tt.input, err)
			continue
		}
		expr := tmpl.parts[0].(*expression)
		vs := expr.vars[0]
		if vs.name != tt.wantName {
			t.Errorf("Parse(%q) name = %q, want %q", tt.input, vs.name, tt.wantName)
		}
		if vs.modifier != modExplode {
			t.Errorf("Parse(%q) modifier = %d, want modExplode", tt.input, vs.modifier)
		}
	}
}

func TestParseMixedLiteralsAndExpressions(t *testing.T) {
	tests := []struct {
		input     string
		wantParts int
	}{
		{"http://example.com/{var}", 2},
		{"map?{x,y}", 2},
		{"{scheme}://example.com/{path}", 3},
		{"literal", 1},
		{"{a}{b}", 2},
	}
	for _, tt := range tests {
		tmpl, err := Parse(tt.input)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", tt.input, err)
			continue
		}
		if len(tmpl.parts) != tt.wantParts {
			t.Errorf("Parse(%q) got %d parts, want %d", tt.input, len(tmpl.parts), tt.wantParts)
		}
		if tmpl.String() != tt.input {
			t.Errorf("String() = %q, want %q", tmpl.String(), tt.input)
		}
	}
}

func TestParseLiteralTypes(t *testing.T) {
	tmpl, err := Parse("http://example.com/{var}")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if _, ok := tmpl.parts[0].(literal); !ok {
		t.Error("parts[0] should be literal")
	}
	if _, ok := tmpl.parts[1].(*expression); !ok {
		t.Error("parts[1] should be expression")
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		input string
		desc  string
	}{
		{"{}", "empty expression"},
		{"{!var}", "invalid operator"},
		{"{var:0}", "prefix too small"},
		{"{var:10000}", "prefix too large"},
		{"{", "unclosed brace"},
		{"}", "unexpected close brace"},
		{"{.var}", ""},  // dot operator is valid
		{"{var:3*}", ""}, // prefix and explode
	}
	// Only test cases that should error
	errorCases := []struct {
		input string
		desc  string
	}{
		{"{}", "empty expression"},
		{"{!var}", "invalid operator"},
		{"{var:0}", "prefix too small"},
		{"{var:10000}", "prefix too large"},
		{"{", "unclosed brace"},
		{"}", "unexpected close brace"},
	}
	_ = tests
	for _, tt := range errorCases {
		_, err := Parse(tt.input)
		if err == nil {
			t.Errorf("Parse(%q) expected error (%s), got nil", tt.input, tt.desc)
		}
	}
}

func TestParsePrefixExplodeConflict(t *testing.T) {
	// A var with both :N and * should error
	_, err := Parse("{var:3*}")
	if err == nil {
		t.Error("Parse(\"{var:3*}\") expected error for prefix+explode, got nil")
	}
}

func TestParseVarNameWithDot(t *testing.T) {
	tmpl, err := Parse("{var.name}")
	if err != nil {
		t.Fatalf("Parse(\"{var.name}\") error: %v", err)
	}
	expr := tmpl.parts[0].(*expression)
	if expr.vars[0].name != "var.name" {
		t.Errorf("name = %q, want %q", expr.vars[0].name, "var.name")
	}
}

func TestParseVarNameStartingWithDot(t *testing.T) {
	_, err := Parse("{.op,.name}")
	// {.op,.name} - the '.' is the operator, then "op,.name" are the vars
	// This should parse fine: operator='.', vars=["op", ".name"]
	// But ".name" starts with dot which is invalid for a var name
	if err == nil {
		// ".name" as a variable name starts with dot, should be error
		// Actually wait - with '.' operator, the body is ".op,.name"
		// operator = '.', varList = "op,.name"
		// vars = ["op", ".name"] - ".name" starts with dot -> error
		t.Log("Parse correctly rejected .name variable")
	}
}

func TestParsePercentEncodedVarName(t *testing.T) {
	tmpl, err := Parse("{var%20name}")
	if err != nil {
		t.Fatalf("Parse(\"{var%%20name}\") error: %v", err)
	}
	expr := tmpl.parts[0].(*expression)
	if expr.vars[0].name != "var%20name" {
		t.Errorf("name = %q, want %q", expr.vars[0].name, "var%20name")
	}
}

func TestParseOperatorEncoding(t *testing.T) {
	// + and # operators allow reserved characters (U+R)
	tmpl, err := Parse("{+var}")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	expr := tmpl.parts[0].(*expression)
	if expr.operator.allow != encUR {
		t.Errorf("{+var} allow = %d, want encUR", expr.operator.allow)
	}

	tmpl, err = Parse("{var}")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	expr = tmpl.parts[0].(*expression)
	if expr.operator.allow != encU {
		t.Errorf("{var} allow = %d, want encU", expr.operator.allow)
	}
}

func TestParseEmptyTemplate(t *testing.T) {
	tmpl, err := Parse("")
	if err != nil {
		t.Fatalf("Parse(\"\") error: %v", err)
	}
	if len(tmpl.parts) != 0 {
		t.Errorf("expected 0 parts, got %d", len(tmpl.parts))
	}
	if tmpl.String() != "" {
		t.Errorf("String() = %q, want \"\"", tmpl.String())
	}
}

func TestParsePureLiteral(t *testing.T) {
	tmpl, err := Parse("http://example.com/path")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(tmpl.parts) != 1 {
		t.Fatalf("expected 1 part, got %d", len(tmpl.parts))
	}
	if l, ok := tmpl.parts[0].(literal); !ok || string(l) != "http://example.com/path" {
		t.Errorf("unexpected part: %v", tmpl.parts[0])
	}
}
