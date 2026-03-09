# uritemplate

A Go implementation of [RFC 6570 URI Templates](https://www.rfc-editor.org/rfc/rfc6570) with zero dependencies outside the standard library.

URI Templates let you describe parameterized URIs and then:

- **Expand** a template with variables to produce a URI
- **Match** a URI against a template to extract variables back out

```go
tmpl := uritemplate.MustParse("https://api.example.com/users/{user}/repos{?sort,page}")

// Expand: template + values → URI
uri, _ := tmpl.Expand(uritemplate.Values{
    "user": uritemplate.String("octocat"),
    "sort": uritemplate.String("updated"),
    "page": uritemplate.String("1"),
})
// "https://api.example.com/users/octocat/repos?sort=updated&page=1"

// Match: template + URI → values
vals, ok := tmpl.Match(uri)
// vals["user"] = "octocat", vals["sort"] = "updated", vals["page"] = "1"
```

## Install

```
go get github.com/dhamidi/uritemplate
```

## Use Cases

- **API clients**: Expand URL templates returned by hypermedia APIs (GitHub, HAL, JSON:API)
- **Routing**: Match incoming request URLs against templates to extract path and query parameters
- **URL building**: Construct URLs safely with proper percent-encoding instead of `fmt.Sprintf`
- **Link generation**: Generate links from templates in server responses

## Operators

RFC 6570 defines operators that control how variables are expanded:

| Expression  | Expansion with `var = "hello"` | Use for                  |
|-------------|-------------------------------|--------------------------|
| `{var}`     | `hello`                       | General string expansion |
| `{+var}`    | `hello`                       | Reserved characters kept |
| `{#var}`    | `#hello`                      | URI fragments            |
| `{.var}`    | `.hello`                      | File extensions, labels  |
| `{/var}`    | `/hello`                      | Path segments            |
| `{;var}`    | `;var=hello`                  | Path parameters          |
| `{?var}`    | `?var=hello`                  | Query strings            |
| `{&var}`    | `&var=hello`                  | Query continuation       |

## Variable Types

```go
// Simple string
uritemplate.String("hello")

// List of strings — expands as comma-separated or exploded
uritemplate.List("red", "green", "blue")

// Key-value pairs — expands as associative array
uritemplate.Keys(
    uritemplate.KeyValue{Key: "width", Value: "100"},
    uritemplate.KeyValue{Key: "height", Value: "200"},
)
```

## Modifiers

```go
// Prefix: truncate to N characters
tmpl := uritemplate.MustParse("{var:3}")
tmpl.Expand(uritemplate.Values{"var": uritemplate.String("hello")})
// → "hel"

// Explode: expand list/keys items separately
tmpl = uritemplate.MustParse("{?list*}")
tmpl.Expand(uritemplate.Values{"list": uritemplate.List("a", "b", "c")})
// → "?list=a&list=b&list=c"
```

## Integration with net/url

The library provides direct integration with Go's `net/url` package:

```go
// Expand into a *url.URL
tmpl := uritemplate.MustParse("https://example.com/users/{user}")
u, err := tmpl.URL(uritemplate.Values{"user": uritemplate.String("octocat")})
// u.Host = "example.com", u.Path = "/users/octocat"

// Extract from a *url.URL (handles query param reordering)
tmpl = uritemplate.MustParse("https://example.com/search{?q,lang}")
u, _ = url.Parse("https://example.com/search?lang=en&q=hello")
vals, ok := tmpl.FromURL(u)
// vals["q"] = "hello", vals["lang"] = "en"
```

## Integration with net/http

A template can act as an HTTP handler, matching incoming requests and extracting variables:

```go
mux := http.NewServeMux()
tmpl := uritemplate.MustParse("/users/{user}/repos{?sort,page}")
mux.Handle("/users/", tmpl.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    vals := uritemplate.ValuesFromRequest(r)
    user := vals["user"]
    // ...
})))
```

When a request URL matches the template, extracted values are stored in the request context. When there is no match, the handler responds with 404 Not Found.

Use `ValuesFromRequest(r)` or `ValuesFromContext(ctx)` to retrieve the extracted values in the inner handler.

## API Overview

| Function / Method     | Description                                         |
|-----------------------|-----------------------------------------------------|
| `Parse(s)`            | Parse a URI template string, returning any errors    |
| `MustParse(s)`        | Like `Parse` but panics on error                     |
| `Template.Expand(v)`  | Expand template with variables into a URI string     |
| `Template.URL(v)`     | Expand into a `*url.URL`                             |
| `Template.Match(s)`   | Extract variables from a URI string                  |
| `Template.FromURL(u)` | Extract variables from a `*url.URL`                  |
| `Template.Handler(h)` | Return an HTTP handler that matches requests         |
| `Template.HandlerFunc(f)` | Convenience wrapper accepting a handler function |
| `Template.String()`   | Return the original template string                  |
| `ValuesFromRequest(r)` | Extract template values from a request context       |
| `ValuesFromContext(ctx)` | Extract template values from a context             |
| `String(s)`           | Create a string variable value                       |
| `List(items...)`      | Create a list variable value                         |
| `Keys(pairs...)`      | Create an associative array variable value           |

## Specification Compliance

This implementation supports all four levels of RFC 6570:

- **Level 1**: Simple string expansion (`{var}`)
- **Level 2**: Reserved (`{+var}`) and fragment (`{#var}`) expansion
- **Level 3**: Multiple variables and all operators (`.`, `/`, `;`, `?`, `&`)
- **Level 4**: Prefix modifier (`{var:3}`) and explode modifier (`{list*}`)

Undefined variables are omitted from the output, following the specification.

## License

See [LICENSE](LICENSE) for details.
