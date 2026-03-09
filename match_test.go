package uritemplate

import (
	"testing"
)

func TestMatch(t *testing.T) {
	tests := []struct {
		name     string
		template string
		uri      string
		wantOK   bool
		wantVals Values
	}{
		{
			name:     "simple variable",
			template: "{var}",
			uri:      "value",
			wantOK:   true,
			wantVals: Values{"var": String("value")},
		},
		{
			name:     "reserved operator",
			template: "{+path}",
			uri:      "/foo/bar",
			wantOK:   true,
			wantVals: Values{"path": String("/foo/bar")},
		},
		{
			name:     "fragment operator",
			template: "{#var}",
			uri:      "#value",
			wantOK:   true,
			wantVals: Values{"var": String("value")},
		},
		{
			name:     "label operator",
			template: "{.var}",
			uri:      ".value",
			wantOK:   true,
			wantVals: Values{"var": String("value")},
		},
		{
			name:     "path operator",
			template: "{/var}",
			uri:      "/value",
			wantOK:   true,
			wantVals: Values{"var": String("value")},
		},
		{
			name:     "path params operator",
			template: "{;x,y}",
			uri:      ";x=1024;y=768",
			wantOK:   true,
			wantVals: Values{"x": String("1024"), "y": String("768")},
		},
		{
			name:     "query operator",
			template: "{?x,y}",
			uri:      "?x=1024&y=768",
			wantOK:   true,
			wantVals: Values{"x": String("1024"), "y": String("768")},
		},
		{
			name:     "query continuation operator",
			template: "{&x}",
			uri:      "&x=1024",
			wantOK:   true,
			wantVals: Values{"x": String("1024")},
		},
		{
			name:     "with literals",
			template: "http://example.com/{var}",
			uri:      "http://example.com/hello",
			wantOK:   true,
			wantVals: Values{"var": String("hello")},
		},
		{
			name:     "explode list named",
			template: "{?list*}",
			uri:      "?list=red&list=green&list=blue",
			wantOK:   true,
			wantVals: Values{"list": List("red", "green", "blue")},
		},
		{
			name:     "no match",
			template: "http://example.com/{var}",
			uri:      "http://other.com/hello",
			wantOK:   false,
		},
		{
			name:     "trailing literal mismatch",
			template: "http://example.com/{var}/details",
			uri:      "http://example.com/hello",
			wantOK:   false,
		},
		{
			name:     "percent encoded value",
			template: "{var}",
			uri:      "Hello%20World%21",
			wantOK:   true,
			wantVals: Values{"var": String("Hello World!")},
		},
		{
			name:     "multiple expressions with literals",
			template: "http://example.com/{var}{?q,lang}",
			uri:      "http://example.com/hello?q=world&lang=en",
			wantOK:   true,
			wantVals: Values{"var": String("hello"), "q": String("world"), "lang": String("en")},
		},
		{
			name:     "path params with empty value",
			template: "{;x,y,empty}",
			uri:      ";x=1024;y=768;empty",
			wantOK:   true,
			wantVals: Values{"x": String("1024"), "y": String("768"), "empty": String("")},
		},
		{
			name:     "query with empty value",
			template: "{?x,y,empty}",
			uri:      "?x=1024&y=768&empty=",
			wantOK:   true,
			wantVals: Values{"x": String("1024"), "y": String("768"), "empty": String("")},
		},
		{
			name:     "reserved with literal suffix",
			template: "{+path}/here",
			uri:      "/foo/bar/here",
			wantOK:   true,
			wantVals: Values{"path": String("/foo/bar")},
		},
		{
			name:     "fragment with literal suffix",
			template: "{#path,x}/here",
			uri:      "#/foo/bar,1024/here",
			wantOK:   true,
			wantVals: Values{"path": String("/foo/bar"), "x": String("1024")},
		},
		{
			name:     "query continuation with literal prefix",
			template: "?fixed=yes{&x}",
			uri:      "?fixed=yes&x=1024",
			wantOK:   true,
			wantVals: Values{"x": String("1024")},
		},
		{
			name:     "label multiple vars",
			template: "X{.x,y}",
			uri:      "X.1024.768",
			wantOK:   true,
			wantVals: Values{"x": String("1024"), "y": String("768")},
		},
		{
			name:     "path multiple vars with literal",
			template: "{/var,x}/here",
			uri:      "/value/1024/here",
			wantOK:   true,
			wantVals: Values{"var": String("value"), "x": String("1024")},
		},
		{
			name:     "explode keys named query",
			template: "{?keys*}",
			uri:      "?semi=%3B&dot=.&comma=%2C",
			wantOK:   true,
			wantVals: Values{"keys": Keys(KeyValue{"semi", ";"}, KeyValue{"dot", "."}, KeyValue{"comma", ","})},
		},
		{
			name:     "explode list unnamed",
			template: "{list*}",
			uri:      "red,green,blue",
			wantOK:   true,
			wantVals: Values{"list": List("red", "green", "blue")},
		},
		{
			name:     "explode keys unnamed",
			template: "{keys*}",
			uri:      "semi=%3B,dot=.,comma=%2C",
			wantOK:   true,
			wantVals: Values{"keys": Keys(KeyValue{"semi", ";"}, KeyValue{"dot", "."}, KeyValue{"comma", ","})},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := Parse(tt.template)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", tt.template, err)
			}
			gotVals, gotOK := tmpl.Match(tt.uri)
			if gotOK != tt.wantOK {
				t.Fatalf("Match(%q) ok = %v, want %v", tt.uri, gotOK, tt.wantOK)
			}
			if !tt.wantOK {
				return
			}
			for name, wantVal := range tt.wantVals {
				gotVal, ok := gotVals[name]
				if !ok {
					t.Errorf("Match(%q) missing variable %q", tt.uri, name)
					continue
				}
				if !valuesEqual(gotVal, wantVal) {
					t.Errorf("Match(%q) variable %q = %v, want %v", tt.uri, name, fmtValue(gotVal), fmtValue(wantVal))
				}
			}
		})
	}
}

func TestMatchRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     Values
	}{
		{
			name:     "simple",
			template: "{var}",
			vars:     Values{"var": String("hello")},
		},
		{
			name:     "with literal",
			template: "http://example.com/{var}",
			vars:     Values{"var": String("hello")},
		},
		{
			name:     "query params",
			template: "http://example.com/{var}{?q,lang}",
			vars:     Values{"var": String("hello"), "q": String("world"), "lang": String("en")},
		},
		{
			name:     "path and query",
			template: "{/path}{?x,y}",
			vars:     Values{"path": String("foo"), "x": String("1"), "y": String("2")},
		},
		{
			name:     "reserved",
			template: "{+path}/here",
			vars:     Values{"path": String("/foo/bar")},
		},
		{
			name:     "fragment",
			template: "{#var}",
			vars:     Values{"var": String("hello")},
		},
		{
			name:     "label",
			template: "X{.var}",
			vars:     Values{"var": String("hello")},
		},
		{
			name:     "path params",
			template: "{;x,y}",
			vars:     Values{"x": String("1024"), "y": String("768")},
		},
		{
			name:     "query continuation",
			template: "?fixed=yes{&x}",
			vars:     Values{"x": String("1024")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := Parse(tt.template)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", tt.template, err)
			}
			uri, err := tmpl.Expand(tt.vars)
			if err != nil {
				t.Fatalf("Expand error: %v", err)
			}
			gotVals, ok := tmpl.Match(uri)
			if !ok {
				t.Fatalf("Match(%q) returned false", uri)
			}
			// Re-expand with matched values and compare
			uri2, err := tmpl.Expand(gotVals)
			if err != nil {
				t.Fatalf("Expand(matched) error: %v", err)
			}
			if uri2 != uri {
				t.Errorf("Round-trip failed: Expand(Match(%q)) = %q", uri, uri2)
			}
		})
	}
}

func valuesEqual(a, b Value) bool {
	if a.kind != b.kind {
		return false
	}
	switch a.kind {
	case kindString:
		return a.str == b.str
	case kindList:
		if len(a.list) != len(b.list) {
			return false
		}
		for i := range a.list {
			if a.list[i] != b.list[i] {
				return false
			}
		}
		return true
	case kindKeys:
		if len(a.keys) != len(b.keys) {
			return false
		}
		for i := range a.keys {
			if a.keys[i] != b.keys[i] {
				return false
			}
		}
		return true
	}
	return false
}

func fmtValue(v Value) string {
	switch v.kind {
	case kindString:
		return "String(" + v.str + ")"
	case kindList:
		s := "List("
		for i, item := range v.list {
			if i > 0 {
				s += ","
			}
			s += item
		}
		return s + ")"
	case kindKeys:
		s := "Keys("
		for i, kv := range v.keys {
			if i > 0 {
				s += ","
			}
			s += kv.Key + "=" + kv.Value
		}
		return s + ")"
	}
	return "?"
}
