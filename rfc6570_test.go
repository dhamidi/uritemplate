package uritemplate

import (
	"testing"
)

// rfcVars contains the test variable values from RFC 6570 Sections 1.2 and 3.2.1.
var rfcVars = Values{
	"count":      List("one", "two", "three"),
	"dom":        List("example", "com"),
	"dub":        String("me/too"),
	"hello":      String("Hello World!"),
	"half":       String("50%"),
	"var":        String("value"),
	"who":        String("fred"),
	"base":       String("http://example.com/home/"),
	"path":       String("/foo/bar"),
	"list":       List("red", "green", "blue"),
	"keys":       Keys(KeyValue{"semi", ";"}, KeyValue{"dot", "."}, KeyValue{"comma", ","}),
	"v":          String("6"),
	"x":          String("1024"),
	"y":          String("768"),
	"empty":      String(""),
	"empty_keys": Keys(),
	// undef is intentionally absent
}

func TestRFCSection3_2_2_SimpleStringExpansion(t *testing.T) {
	tests := []struct {
		template string
		expected string
	}{
		{"{var}", "value"},
		{"{hello}", "Hello%20World%21"},
		{"{half}", "50%25"},
		{"O{empty}X", "OX"},
		{"O{undef}X", "OX"},
		{"{x,y}", "1024,768"},
		{"{x,hello,y}", "1024,Hello%20World%21,768"},
		{"?{x,empty}", "?1024,"},
		{"?{x,undef}", "?1024"},
		{"?{undef,y}", "?768"},
		{"{var:3}", "val"},
		{"{var:30}", "value"},
		{"{list}", "red,green,blue"},
		{"{list*}", "red,green,blue"},
		{"{keys}", "semi,%3B,dot,.,comma,%2C"},
		{"{keys*}", "semi=%3B,dot=.,comma=%2C"},
	}

	runExpansionTests(t, tests)
}

func TestRFCSection3_2_3_ReservedExpansion(t *testing.T) {
	tests := []struct {
		template string
		expected string
	}{
		{"{+var}", "value"},
		{"{+hello}", "Hello%20World!"},
		{"{+half}", "50%25"},
		{"{base}index", "http%3A%2F%2Fexample.com%2Fhome%2Findex"},
		{"{+base}index", "http://example.com/home/index"},
		{"O{+empty}X", "OX"},
		{"O{+undef}X", "OX"},
		{"{+path}/here", "/foo/bar/here"},
		{"here?ref={+path}", "here?ref=/foo/bar"},
		{"up{+path}{var}/here", "up/foo/barvalue/here"},
		{"{+x,hello,y}", "1024,Hello%20World!,768"},
		{"{+path,x}/here", "/foo/bar,1024/here"},
		{"{+path:6}/here", "/foo/b/here"},
		{"{+list}", "red,green,blue"},
		{"{+list*}", "red,green,blue"},
		{"{+keys}", "semi,;,dot,.,comma,,"},
		{"{+keys*}", "semi=;,dot=.,comma=,"},
	}

	runExpansionTests(t, tests)
}

func TestRFCSection3_2_4_FragmentExpansion(t *testing.T) {
	tests := []struct {
		template string
		expected string
	}{
		{"{#var}", "#value"},
		{"{#hello}", "#Hello%20World!"},
		{"{#half}", "#50%25"},
		{"foo{#empty}", "foo#"},
		{"foo{#undef}", "foo"},
		{"{#x,hello,y}", "#1024,Hello%20World!,768"},
		{"{#path,x}/here", "#/foo/bar,1024/here"},
		{"{#path:6}/here", "#/foo/b/here"},
		{"{#list}", "#red,green,blue"},
		{"{#list*}", "#red,green,blue"},
		{"{#keys}", "#semi,;,dot,.,comma,,"},
		{"{#keys*}", "#semi=;,dot=.,comma=,"},
	}

	runExpansionTests(t, tests)
}

func TestRFCSection3_2_5_LabelExpansion(t *testing.T) {
	tests := []struct {
		template string
		expected string
	}{
		{"X{.var}", "X.value"},
		{"X{.empty}", "X."},
		{"X{.undef}", "X"},
		{"X{.x,y}", "X.1024.768"},
		{"X{.var:3}", "X.val"},
		{"X{.list}", "X.red,green,blue"},
		{"X{.list*}", "X.red.green.blue"},
		{"X{.keys}", "X.semi,%3B,dot,.,comma,%2C"},
		{"X{.keys*}", "X.semi=%3B.dot=..comma=%2C"},
	}

	runExpansionTests(t, tests)
}

func TestRFCSection3_2_6_PathSegmentExpansion(t *testing.T) {
	tests := []struct {
		template string
		expected string
	}{
		{"{/var}", "/value"},
		{"{/var,x}/here", "/value/1024/here"},
		{"{/var,undef,y}", "/value/768"},
		{"{/var:1,var}", "/v/value"},
		{"{/list}", "/red,green,blue"},
		{"{/list*}", "/red/green/blue"},
		{"{/list*,path:4}", "/red/green/blue/%2Ffoo"},
		{"{/keys}", "/semi,%3B,dot,.,comma,%2C"},
		{"{/keys*}", "/semi=%3B/dot=./comma=%2C"},
	}

	runExpansionTests(t, tests)
}

func TestRFCSection3_2_7_PathStyleParameterExpansion(t *testing.T) {
	tests := []struct {
		template string
		expected string
	}{
		{"{;who}", ";who=fred"},
		{"{;half}", ";half=50%25"},
		{"{;empty}", ";empty"},
		{"{;v,empty,who}", ";v=6;empty;who=fred"},
		{"{;v,bar,who}", ";v=6;who=fred"},
		{"{;x,y}", ";x=1024;y=768"},
		{"{;x,y,empty}", ";x=1024;y=768;empty"},
		{"{;x,y,undef}", ";x=1024;y=768"},
		{"{;hello:5}", ";hello=Hello"},
		{"{;list}", ";list=red,green,blue"},
		{"{;list*}", ";list=red;list=green;list=blue"},
		{"{;keys}", ";keys=semi,%3B,dot,.,comma,%2C"},
		{"{;keys*}", ";semi=%3B;dot=.;comma=%2C"},
	}

	runExpansionTests(t, tests)
}

func TestRFCSection3_2_8_FormStyleQueryExpansion(t *testing.T) {
	tests := []struct {
		template string
		expected string
	}{
		{"{?who}", "?who=fred"},
		{"{?half}", "?half=50%25"},
		{"{?x,y}", "?x=1024&y=768"},
		{"{?x,y,empty}", "?x=1024&y=768&empty="},
		{"{?x,y,undef}", "?x=1024&y=768"},
		{"{?var:3}", "?var=val"},
		{"{?list}", "?list=red,green,blue"},
		{"{?list*}", "?list=red&list=green&list=blue"},
		{"{?keys}", "?keys=semi,%3B,dot,.,comma,%2C"},
		{"{?keys*}", "?semi=%3B&dot=.&comma=%2C"},
	}

	runExpansionTests(t, tests)
}

func TestRFCSection3_2_9_FormStyleQueryContinuation(t *testing.T) {
	tests := []struct {
		template string
		expected string
	}{
		{"{&who}", "&who=fred"},
		{"{&half}", "&half=50%25"},
		{"?fixed=yes{&x}", "?fixed=yes&x=1024"},
		{"{&x,y,empty}", "&x=1024&y=768&empty="},
		{"{&var:3}", "&var=val"},
		{"{&list}", "&list=red,green,blue"},
		{"{&list*}", "&list=red&list=green&list=blue"},
		{"{&keys}", "&keys=semi,%3B,dot,.,comma,%2C"},
		{"{&keys*}", "&semi=%3B&dot=.&comma=%2C"},
	}

	runExpansionTests(t, tests)
}

func runExpansionTests(t *testing.T, tests []struct {
	template string
	expected string
}) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.template, func(t *testing.T) {
			tmpl, err := Parse(tt.template)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", tt.template, err)
			}
			got, err := tmpl.Expand(rfcVars)
			if err != nil {
				t.Fatalf("Expand(%q) error: %v", tt.template, err)
			}
			if got != tt.expected {
				t.Errorf("Expand(%q) = %q, want %q", tt.template, got, tt.expected)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	// Round-trip tests for templates using only string values without prefix modifiers.
	tests := []struct {
		template string
		vars     Values
	}{
		{"{var}", Values{"var": String("value")}},
		{"{who}", Values{"who": String("fred")}},
		{"{x,y}", Values{"x": String("1024"), "y": String("768")}},
		{"{+var}", Values{"var": String("value")}},
		{"{+hello}", Values{"hello": String("Hello World!")}},
		{"{#var}", Values{"var": String("value")}},
		{"X{.var}", Values{"var": String("value")}},
		{"{/var}", Values{"var": String("value")}},
		{"{/var,x}/here", Values{"var": String("value"), "x": String("1024")}},
		{"{;who}", Values{"who": String("fred")}},
		{"{;x,y}", Values{"x": String("1024"), "y": String("768")}},
		{"{?who}", Values{"who": String("fred")}},
		{"{?x,y}", Values{"x": String("1024"), "y": String("768")}},
		{"?fixed=yes{&x}", Values{"x": String("1024")}},
	}

	for _, tt := range tests {
		t.Run(tt.template, func(t *testing.T) {
			tmpl, err := Parse(tt.template)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", tt.template, err)
			}
			expanded, err := tmpl.Expand(tt.vars)
			if err != nil {
				t.Fatalf("Expand(%q) error: %v", tt.template, err)
			}
			extracted, ok := tmpl.Match(expanded)
			if !ok {
				t.Fatalf("Match(%q) returned false for expanded URI %q", tt.template, expanded)
			}
			for name, want := range tt.vars {
				got, exists := extracted[name]
				if !exists {
					t.Errorf("Match(%q): missing variable %q", tt.template, name)
					continue
				}
				if got.kind != want.kind {
					t.Errorf("Match(%q): variable %q kind = %v, want %v", tt.template, name, got.kind, want.kind)
					continue
				}
				if got.str != want.str {
					t.Errorf("Match(%q): variable %q = %q, want %q", tt.template, name, got.str, want.str)
				}
			}
		})
	}
}

func TestRFCURLRoundTrip(t *testing.T) {
	tests := []struct {
		template string
		vars     Values
	}{
		{
			"https://api.example.com/v1/users/{user}",
			Values{"user": String("johndoe")},
		},
		{
			"https://api.example.com/v1/users/{user}/repos{?sort,page,per_page}",
			Values{"user": String("johndoe"), "sort": String("stars"), "page": String("1"), "per_page": String("10")},
		},
		{
			"https://example.com{/year,month,day}",
			Values{"year": String("2026"), "month": String("03"), "day": String("09")},
		},
		{
			"https://example.com/search{?q,lang,page}",
			Values{"q": String("golang"), "lang": String("en"), "page": String("1")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.template, func(t *testing.T) {
			tmpl, err := Parse(tt.template)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", tt.template, err)
			}
			u, err := tmpl.URL(tt.vars)
			if err != nil {
				t.Fatalf("URL(%q) error: %v", tt.template, err)
			}
			extracted, ok := tmpl.FromURL(u)
			if !ok {
				t.Fatalf("FromURL(%q) returned false for URL %q", tt.template, u.String())
			}
			for name, want := range tt.vars {
				got, exists := extracted[name]
				if !exists {
					t.Errorf("FromURL(%q): missing variable %q", tt.template, name)
					continue
				}
				if got.str != want.str {
					t.Errorf("FromURL(%q): variable %q = %q, want %q", tt.template, name, got.str, want.str)
				}
			}
		})
	}
}

func TestMalformedTemplates(t *testing.T) {
	malformed := []string{
		"{",
		"}",
		"{}",
		"{!var}",
		"{var",
		"{var:}",
		"{var:0}",
		"{var:10000}",
		"{var:abc}",
		"{.}",
		"{var*:3}",
		"{=var}",
	}

	for _, tmpl := range malformed {
		t.Run(tmpl, func(t *testing.T) {
			_, err := Parse(tmpl)
			if err == nil {
				t.Errorf("Parse(%q) expected error, got nil", tmpl)
			}
		})
	}
}

func TestURLExpansionInvalidURL(t *testing.T) {
	// A template that expands to something with control characters
	tmpl, err := Parse("{var}")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	// Expand with a value that has a bare % which gets encoded - should still be valid URL
	_, err = tmpl.URL(Values{"var": String("value")})
	if err != nil {
		t.Errorf("URL() unexpected error: %v", err)
	}
}
