package uritemplate_test

import (
	"fmt"
	"net/url"

	"github.com/dhamidi/uritemplate"
)

func Example() {
	tmpl, err := uritemplate.Parse("https://api.example.com/users/{user}/repos{?sort,page}")
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	uri, err := tmpl.Expand(uritemplate.Values{
		"user": uritemplate.String("octocat"),
		"sort": uritemplate.String("updated"),
		"page": uritemplate.String("1"),
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println(uri)
	// Output: https://api.example.com/users/octocat/repos?sort=updated&page=1
}

func ExampleParse() {
	tmpl, err := uritemplate.Parse("https://example.com/{section}/{id}")
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println(tmpl.String())
	// Output: https://example.com/{section}/{id}
}

func ExampleMustParse() {
	tmpl := uritemplate.MustParse("/users/{user}")
	fmt.Println(tmpl.String())
	// Output: /users/{user}
}

func ExampleTemplate_Expand() {
	tmpl := uritemplate.MustParse("{/who,dub}")
	uri, err := tmpl.Expand(uritemplate.Values{
		"who": uritemplate.String("fred"),
		"dub": uritemplate.String("me/too"),
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println(uri)
	// Output: /fred/me%2Ftoo
}

func ExampleTemplate_Expand_query() {
	tmpl := uritemplate.MustParse("https://example.com/search{?q,lang}")
	uri, err := tmpl.Expand(uritemplate.Values{
		"q":    uritemplate.String("URI Templates"),
		"lang": uritemplate.String("en"),
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println(uri)
	// Output: https://example.com/search?q=URI%20Templates&lang=en
}

func ExampleTemplate_Expand_list() {
	tmpl := uritemplate.MustParse("{?list*}")
	uri, err := tmpl.Expand(uritemplate.Values{
		"list": uritemplate.List("red", "green", "blue"),
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println(uri)
	// Output: ?list=red&list=green&list=blue
}

func ExampleTemplate_Expand_keys() {
	tmpl := uritemplate.MustParse("{?keys*}")
	uri, err := tmpl.Expand(uritemplate.Values{
		"keys": uritemplate.Keys(
			uritemplate.KeyValue{Key: "semi", Value: ";"},
			uritemplate.KeyValue{Key: "dot", Value: "."},
		),
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println(uri)
	// Output: ?semi=%3B&dot=.
}

func ExampleTemplate_Expand_reserved() {
	tmpl := uritemplate.MustParse("{+base}{+path}/here")
	uri, err := tmpl.Expand(uritemplate.Values{
		"base": uritemplate.String("http://example.com/home/"),
		"path": uritemplate.String("/foo/bar"),
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println(uri)
	// Output: http://example.com/home//foo/bar/here
}

func ExampleTemplate_Expand_fragment() {
	tmpl := uritemplate.MustParse("https://example.com/page{#section}")
	uri, err := tmpl.Expand(uritemplate.Values{
		"section": uritemplate.String("details"),
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println(uri)
	// Output: https://example.com/page#details
}

func ExampleTemplate_Expand_prefix() {
	tmpl := uritemplate.MustParse("{var:3}")
	uri, err := tmpl.Expand(uritemplate.Values{
		"var": uritemplate.String("value"),
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println(uri)
	// Output: val
}

func ExampleTemplate_Expand_undefinedVariables() {
	tmpl := uritemplate.MustParse("https://example.com/search{?q,lang}")
	uri, err := tmpl.Expand(uritemplate.Values{
		"q": uritemplate.String("hello"),
		// "lang" is not provided — it will be omitted
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println(uri)
	// Output: https://example.com/search?q=hello
}

func ExampleTemplate_URL() {
	tmpl := uritemplate.MustParse("https://api.example.com/users/{user}")
	u, err := tmpl.URL(uritemplate.Values{
		"user": uritemplate.String("octocat"),
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println(u.Host)
	fmt.Println(u.Path)
	// Output:
	// api.example.com
	// /users/octocat
}

func ExampleTemplate_Match() {
	tmpl := uritemplate.MustParse("https://example.com/users/{user}/repos/{repo}")
	vals, ok := tmpl.Match("https://example.com/users/octocat/repos/hello-world")
	if !ok {
		fmt.Println("no match")
		return
	}

	// Re-expand to verify the round-trip
	uri, _ := tmpl.Expand(vals)
	fmt.Println(uri)
	// Output: https://example.com/users/octocat/repos/hello-world
}

func ExampleTemplate_FromURL() {
	tmpl := uritemplate.MustParse("https://example.com/search{?q,lang}")
	u, _ := url.Parse("https://example.com/search?lang=en&q=hello")
	vals, ok := tmpl.FromURL(u)
	if !ok {
		fmt.Println("no match")
		return
	}

	// Re-expand to verify the round-trip — note query order matches the template
	uri, _ := tmpl.Expand(vals)
	fmt.Println(uri)
	// Output: https://example.com/search?q=hello&lang=en
}

func ExampleString() {
	v := uritemplate.String("hello")
	vals := uritemplate.Values{"greeting": v}

	tmpl := uritemplate.MustParse("{greeting}")
	uri, _ := tmpl.Expand(vals)
	fmt.Println(uri)
	// Output: hello
}

func ExampleList() {
	v := uritemplate.List("red", "green", "blue")
	vals := uritemplate.Values{"colors": v}

	tmpl := uritemplate.MustParse("{colors}")
	uri, _ := tmpl.Expand(vals)
	fmt.Println(uri)
	// Output: red,green,blue
}

func ExampleKeys() {
	v := uritemplate.Keys(
		uritemplate.KeyValue{Key: "width", Value: "100"},
		uritemplate.KeyValue{Key: "height", Value: "200"},
	)
	vals := uritemplate.Values{"size": v}

	tmpl := uritemplate.MustParse("{?size*}")
	uri, _ := tmpl.Expand(vals)
	fmt.Println(uri)
	// Output: ?width=100&height=200
}

func ExampleTemplate_String() {
	tmpl := uritemplate.MustParse("{+base}{/path*}")
	fmt.Println(tmpl.String())
	// Output: {+base}{/path*}
}
