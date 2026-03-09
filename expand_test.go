package uritemplate

import "testing"

// RFC 6570 Section 3 test variables
var testVars = Values{
	"count":    List("one", "two", "three"),
	"dom":      List("example", "com"),
	"dub":      String("me/too"),
	"hello":    String("Hello World!"),
	"half":     String("50%"),
	"var":      String("value"),
	"who":      String("fred"),
	"base":     String("http://example.com/home/"),
	"path":     String("/foo/bar"),
	"list":     List("red", "green", "blue"),
	"keys":     Keys(KeyValue{"semi", ";"}, KeyValue{"dot", "."}, KeyValue{"comma", ","}),
	"v":        String("6"),
	"x":        String("1024"),
	"y":        String("768"),
	"empty":    String(""),
	"empty_keys": Keys(),
}

func TestExpand(t *testing.T) {
	tests := []struct {
		name     string
		template string
		want     string
	}{
		// Level 1 (Section 3.2.2)
		{"L1 var", "{var}", "value"},
		{"L1 hello", "{hello}", "Hello%20World%21"},

		// Level 2 (Section 3.2.3)
		{"L2 +var", "{+var}", "value"},
		{"L2 +hello", "{+hello}", "Hello%20World!"},
		{"L2 +half", "{+half}", "50%25"},
		{"L2 +base+path", "{+base}{+path}/here", "http://example.com/home//foo/bar/here"},
		{"L2 #var", "{#var}", "#value"},
		{"L2 #hello", "{#hello}", "#Hello%20World!"},

		// Level 3 (Section 3.2.4-3.2.9)
		// multiple variables
		{"L3 map", "map?{x,y}", "map?1024,768"},
		{"L3 x,hello,y", "{x,hello,y}", "1024,Hello%20World%21,768"},

		// + operator
		{"L3 +x,hello,y", "{+x,hello,y}", "1024,Hello%20World!,768"},
		{"L3 +path,x", "{+path,x}/here", "/foo/bar,1024/here"},

		// # operator
		{"L3 #x,hello,y", "{#x,hello,y}", "#1024,Hello%20World!,768"},
		{"L3 #path,x", "{#path,x}/here", "#/foo/bar,1024/here"},

		// . operator
		{"L3 X.var", "X{.var}", "X.value"},
		{"L3 X.x,y", "X{.x,y}", "X.1024.768"},

		// / operator
		{"L3 /var", "{/var}", "/value"},
		{"L3 /var,x", "{/var,x}/here", "/value/1024/here"},

		// ; operator
		{"L3 ;x,y", "{;x,y}", ";x=1024;y=768"},
		{"L3 ;x,y,empty", "{;x,y,empty}", ";x=1024;y=768;empty"},

		// ? operator
		{"L3 ?x,y", "{?x,y}", "?x=1024&y=768"},
		{"L3 ?x,y,empty", "{?x,y,empty}", "?x=1024&y=768&empty="},

		// & operator
		{"L3 &x,y,empty", "?fixed=yes{&x}", "?fixed=yes&x=1024"},

		// Level 4 - prefix
		{"L4 var:3", "{var:3}", "val"},
		{"L4 var:30", "{var:30}", "value"},
		{"L4 +var:3", "{+var:3}", "val"},

		// Level 4 - list without explode
		{"L4 list", "{list}", "red,green,blue"},
		{"L4 +list", "{+list}", "red,green,blue"},
		{"L4 #list", "{#list}", "#red,green,blue"},

		// Level 4 - list with explode
		{"L4 list*", "{list*}", "red,green,blue"},
		{"L4 +list*", "{+list*}", "red,green,blue"},
		{"L4 #list*", "{#list*}", "#red,green,blue"},
		{"L4 .list", "X{.list}", "X.red,green,blue"},
		{"L4 .list*", "X{.list*}", "X.red.green.blue"},
		{"L4 /list", "{/list}", "/red,green,blue"},
		{"L4 /list*", "{/list*}", "/red/green/blue"},
		{"L4 ;list", "{;list}", ";list=red,green,blue"},
		{"L4 ;list*", "{;list*}", ";list=red;list=green;list=blue"},
		{"L4 ?list", "{?list}", "?list=red,green,blue"},
		{"L4 ?list*", "{?list*}", "?list=red&list=green&list=blue"},
		{"L4 &list*", "{&list*}", "&list=red&list=green&list=blue"},

		// Level 4 - keys without explode
		{"L4 keys", "{keys}", "semi,%3B,dot,.,comma,%2C"},
		{"L4 +keys", "{+keys}", "semi,;,dot,.,comma,,"},
		{"L4 #keys", "{#keys}", "#semi,;,dot,.,comma,,"},

		// Level 4 - keys with explode
		{"L4 keys*", "{keys*}", "semi=%3B,dot=.,comma=%2C"},
		{"L4 +keys*", "{+keys*}", "semi=;,dot=.,comma=,"},
		{"L4 #keys*", "{#keys*}", "#semi=;,dot=.,comma=,"},
		{"L4 .keys*", "X{.keys*}", "X.semi=%3B.dot=..comma=%2C"},
		{"L4 /keys*", "{/keys*}", "/semi=%3B/dot=./comma=%2C"},
		{"L4 ;keys*", "{;keys*}", ";semi=%3B;dot=.;comma=%2C"},
		{"L4 ?keys*", "{?keys*}", "?semi=%3B&dot=.&comma=%2C"},
		{"L4 &keys*", "{&keys*}", "&semi=%3B&dot=.&comma=%2C"},

		// Undefined variables
		{"undefined single", "{undef}", ""},
		{"undefined in list", "{undef,var}", "value"},
		{"undefined all", "{undef,undef2}", ""},
		{"undefined +", "{+undef}", ""},
		{"undefined #", "{#undef}", ""},
		{"undefined ?", "{?undef}", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := Parse(tt.template)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", tt.template, err)
			}
			got, err := tmpl.Expand(testVars)
			if err != nil {
				t.Fatalf("Expand(%q) error: %v", tt.template, err)
			}
			if got != tt.want {
				t.Errorf("Expand(%q) = %q, want %q", tt.template, got, tt.want)
			}
		})
	}
}
