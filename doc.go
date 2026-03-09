// Package uritemplate implements RFC 6570 URI Templates for Go.
//
// URI Templates provide a way to describe a range of URIs through variable
// expansion. They are used extensively in hypermedia APIs (e.g. GitHub, HAL,
// JSON:API) to express parameterized URLs without string concatenation.
//
// This package has zero dependencies outside the Go standard library.
//
// # Getting Started
//
// A URI template is a string like "https://api.example.com/users/{user}/repos{?sort,page}".
// The parts in curly braces are expressions that get expanded with variable values.
//
// The basic workflow is: parse a template, then expand it with values.
//
//	tmpl, err := uritemplate.Parse("https://api.example.com/users/{user}/repos{?sort,page}")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	uri, err := tmpl.Expand(uritemplate.Values{
//		"user": uritemplate.String("octocat"),
//		"sort": uritemplate.String("updated"),
//		"page": uritemplate.String("1"),
//	})
//	// uri = "https://api.example.com/users/octocat/repos?sort=updated&page=1"
//
// # Working with net/url
//
// Use [Template.URL] to expand a template directly into a [net/url.URL]:
//
//	tmpl := uritemplate.MustParse("https://api.example.com/users/{user}")
//	u, err := tmpl.URL(uritemplate.Values{"user": uritemplate.String("octocat")})
//	// u.Host = "api.example.com", u.Path = "/users/octocat"
//
// Use [Template.FromURL] to extract variable values from a [net/url.URL] back
// into template variables. This is particularly useful for routing:
//
//	tmpl := uritemplate.MustParse("https://api.example.com/users/{user}/repos{?sort}")
//	u, _ := url.Parse("https://api.example.com/users/octocat/repos?sort=updated")
//	vals, ok := tmpl.FromURL(u)
//	// vals["user"] = String("octocat"), vals["sort"] = String("updated")
//
// # Working with net/http
//
// A [Template] can act as an HTTP handler that dispatches to another handler
// when the request URL matches the template. This integrates naturally with
// [net/http.ServeMux]:
//
//	mux := http.NewServeMux()
//	tmpl := uritemplate.MustParse("/users/{user}/repos{?sort,page}")
//	mux.Handle("/users/", tmpl.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//	    vals := uritemplate.ValuesFromRequest(r)
//	    user := vals["user"]
//	    // ...
//	})))
//
// The handler matches incoming request URLs against the template pattern.
// When a request matches, the extracted variable values are stored in the
// request context and the inner handler is called. When there is no match,
// the handler responds with 404 Not Found.
//
// Use [ValuesFromRequest] or [ValuesFromContext] to retrieve extracted values
// in the inner handler:
//
//	vals := uritemplate.ValuesFromRequest(r)
//	user := vals["user"]  // uritemplate.Value
//
// # Variable Types
//
// RFC 6570 supports three types of variable values:
//
//   - [String] — a simple string value
//   - [List] — an ordered list of strings
//   - [Keys] — an ordered list of key-value pairs (associative array)
//
// These are collected in a [Values] map that is passed to [Template.Expand]:
//
//	vals := uritemplate.Values{
//		"name":   uritemplate.String("ferret"),
//		"colors": uritemplate.List("red", "green", "blue"),
//		"meta":   uritemplate.Keys(
//			uritemplate.KeyValue{Key: "semi", Value: ";"},
//			uritemplate.KeyValue{Key: "dot", Value: "."},
//		),
//	}
//
// Undefined variables (absent from the Values map) are silently omitted during
// expansion, following the RFC 6570 specification.
//
// # Operators
//
// Expressions support operators that control how variables are expanded.
// Each operator changes the prefix, separator, and encoding behavior:
//
//	{var}       — Simple expansion: "value"
//	{+var}      — Reserved expansion (preserves /, :, etc.): "/foo/bar"
//	{#var}      — Fragment expansion: "#value"
//	{.var}      — Label expansion: ".value"
//	{/var}      — Path segment expansion: "/value"
//	{;var}      — Path parameter expansion: ";var=value"
//	{?var}      — Query expansion: "?var=value"
//	{&var}      — Query continuation: "&var=value"
//
// # Modifiers
//
// Variables can have modifiers:
//
//	{var:3}     — Prefix modifier: expand only the first 3 characters
//	{list*}     — Explode modifier: expand each list item or key-value pair separately
//
// For example, given list = ["red","green","blue"]:
//
//	{list}   → "red,green,blue"
//	{list*}  → "red,green,blue"
//	{?list*} → "?list=red&list=green&list=blue"
//
// # Matching and Extraction
//
// Besides expanding templates into URIs, this package can perform the reverse
// operation: extracting variable values from a URI that matches a template.
//
// Use [Template.Match] for raw URI strings:
//
//	tmpl := uritemplate.MustParse("http://example.com/{section}/{id}")
//	vals, ok := tmpl.Match("http://example.com/users/42")
//	// vals["section"] = String("users"), vals["id"] = String("42")
//
// Use [Template.FromURL] for [net/url.URL] values. It handles query parameter
// reordering automatically.
//
// # Thread Safety
//
// A parsed [Template] is safe for concurrent use. [Parse] and [MustParse]
// create immutable template values; all methods on [Template] are read-only.
package uritemplate
