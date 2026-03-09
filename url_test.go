package uritemplate

import (
	"net/url"
	"testing"
)

func TestMustParse(t *testing.T) {
	// Valid template should not panic
	tmpl := MustParse("{var}")
	if tmpl == nil {
		t.Fatal("MustParse returned nil")
	}
}

func TestMustParsePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("MustParse did not panic on invalid template")
		}
	}()
	MustParse("{")
}

func TestTemplateURL(t *testing.T) {
	tests := []struct {
		name       string
		template   string
		vars       Values
		wantScheme string
		wantHost   string
		wantPath   string
		wantQuery  url.Values
		wantFrag   string
		wantErr    bool
	}{
		{
			name:       "simple path template",
			template:   "https://example.com/users/{user}",
			vars:       Values{"user": String("octocat")},
			wantScheme: "https",
			wantHost:   "example.com",
			wantPath:   "/users/octocat",
		},
		{
			name:       "template with query params",
			template:   "https://example.com/search{?q,lang}",
			vars:       Values{"q": String("hello"), "lang": String("en")},
			wantScheme: "https",
			wantHost:   "example.com",
			wantPath:   "/search",
			wantQuery:  url.Values{"q": {"hello"}, "lang": {"en"}},
		},
		{
			name:       "template with fragment",
			template:   "https://example.com/page{#section}",
			vars:       Values{"section": String("intro")},
			wantScheme: "https",
			wantHost:   "example.com",
			wantPath:   "/page",
			wantFrag:   "intro",
		},
		{
			name:       "template with path segments",
			template:   "https://example.com{/a,b}",
			vars:       Values{"a": String("foo"), "b": String("bar")},
			wantScheme: "https",
			wantHost:   "example.com",
			wantPath:   "/foo/bar",
		},
		{
			name:       "undefined variables omitted",
			template:   "https://example.com/search{?q,lang}",
			vars:       Values{"q": String("hello")},
			wantScheme: "https",
			wantHost:   "example.com",
			wantPath:   "/search",
			wantQuery:  url.Values{"q": {"hello"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := MustParse(tt.template)
			u, err := tmpl.URL(tt.vars)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("URL() error: %v", err)
			}
			if tt.wantScheme != "" && u.Scheme != tt.wantScheme {
				t.Errorf("Scheme = %q, want %q", u.Scheme, tt.wantScheme)
			}
			if tt.wantHost != "" && u.Host != tt.wantHost {
				t.Errorf("Host = %q, want %q", u.Host, tt.wantHost)
			}
			if tt.wantPath != "" && u.Path != tt.wantPath {
				t.Errorf("Path = %q, want %q", u.Path, tt.wantPath)
			}
			if tt.wantQuery != nil {
				got := u.Query()
				for k, wantVals := range tt.wantQuery {
					gotVals := got[k]
					if len(gotVals) != len(wantVals) {
						t.Errorf("Query[%q] = %v, want %v", k, gotVals, wantVals)
						continue
					}
					for i := range wantVals {
						if gotVals[i] != wantVals[i] {
							t.Errorf("Query[%q][%d] = %q, want %q", k, i, gotVals[i], wantVals[i])
						}
					}
				}
			}
			if tt.wantFrag != "" && u.Fragment != tt.wantFrag {
				t.Errorf("Fragment = %q, want %q", u.Fragment, tt.wantFrag)
			}
		})
	}
}

func TestFromURL(t *testing.T) {
	tests := []struct {
		name     string
		template string
		url      string
		wantOK   bool
		wantVals Values
	}{
		{
			name:     "basic path variable",
			template: "https://example.com/users/{user}",
			url:      "https://example.com/users/torvalds",
			wantOK:   true,
			wantVals: Values{"user": String("torvalds")},
		},
		{
			name:     "query params in different order",
			template: "https://example.com/search{?q,lang}",
			url:      "https://example.com/search?lang=en&q=hello",
			wantOK:   true,
			wantVals: Values{"q": String("hello"), "lang": String("en")},
		},
		{
			name:     "URL with fragment",
			template: "https://example.com/page{#section}",
			url:      "https://example.com/page#intro",
			wantOK:   true,
			wantVals: Values{"section": String("intro")},
		},
		{
			name:     "missing optional query params",
			template: "https://example.com/search{?q,lang,page}",
			url:      "https://example.com/search?q=hello",
			wantOK:   true,
			wantVals: Values{"q": String("hello")},
		},
		{
			name:     "no match",
			template: "https://example.com/users/{user}",
			url:      "https://other.com/users/torvalds",
			wantOK:   false,
		},
		{
			name:     "query continuation different order",
			template: "https://example.com/search?fixed=yes{&q,lang}",
			url:      "https://example.com/search?fixed=yes&lang=en&q=hello",
			wantOK:   true,
			wantVals: Values{"q": String("hello"), "lang": String("en")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := MustParse(tt.template)
			u, err := url.Parse(tt.url)
			if err != nil {
				t.Fatalf("url.Parse(%q) error: %v", tt.url, err)
			}
			gotVals, gotOK := tmpl.FromURL(u)
			if gotOK != tt.wantOK {
				t.Fatalf("FromURL() ok = %v, want %v", gotOK, tt.wantOK)
			}
			if !tt.wantOK {
				return
			}
			for name, wantVal := range tt.wantVals {
				gotVal, ok := gotVals[name]
				if !ok {
					t.Errorf("missing variable %q", name)
					continue
				}
				if !valuesEqual(gotVal, wantVal) {
					t.Errorf("variable %q = %v, want %v", name, fmtValue(gotVal), fmtValue(wantVal))
				}
			}
			// Check no extra query vars present that shouldn't be
			for name := range tt.wantVals {
				if _, ok := gotVals[name]; !ok {
					t.Errorf("missing expected variable %q", name)
				}
			}
		})
	}
}

func TestURLRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     Values
	}{
		{
			name:     "simple path",
			template: "https://example.com/users/{user}",
			vars:     Values{"user": String("octocat")},
		},
		{
			name:     "path and query",
			template: "https://api.example.com/users/{user}/repos{?sort,page}",
			vars:     Values{"user": String("octocat"), "sort": String("updated"), "page": String("1")},
		},
		{
			name:     "fragment",
			template: "https://example.com/page{#section}",
			vars:     Values{"section": String("intro")},
		},
		{
			name:     "path segments",
			template: "https://example.com{/a,b}",
			vars:     Values{"a": String("foo"), "b": String("bar")},
		},
		{
			name:     "query continuation",
			template: "https://example.com/search?fixed=yes{&q,lang}",
			vars:     Values{"q": String("hello"), "lang": String("en")},
		},
		{
			name:     "label operator",
			template: "https://example.com/X{.ext}",
			vars:     Values{"ext": String("json")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := MustParse(tt.template)
			u, err := tmpl.URL(tt.vars)
			if err != nil {
				t.Fatalf("URL() error: %v", err)
			}
			extracted, ok := tmpl.FromURL(u)
			if !ok {
				t.Fatalf("FromURL(%q) returned false", u.String())
			}
			for name, wantVal := range tt.vars {
				gotVal, exists := extracted[name]
				if !exists {
					t.Errorf("round-trip missing variable %q", name)
					continue
				}
				if !valuesEqual(gotVal, wantVal) {
					t.Errorf("round-trip variable %q = %v, want %v", name, fmtValue(gotVal), fmtValue(wantVal))
				}
			}
		})
	}
}
