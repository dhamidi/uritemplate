package uritemplate

import (
	"context"
	"net/http"
)

// contextKey is an unexported type for context keys to avoid collisions.
type contextKey struct{}

// valuesKey is the context key for storing extracted template values.
var valuesKey = contextKey{}

// ValuesFromContext returns the URI template values stored in the context
// by a template handler. Returns nil if no values are present.
func ValuesFromContext(ctx context.Context) Values {
	vals, _ := ctx.Value(valuesKey).(Values)
	return vals
}

// ValuesFromRequest is a convenience function that extracts URI template
// values from the request's context. Equivalent to
// ValuesFromContext(r.Context()).
func ValuesFromRequest(r *http.Request) Values {
	return ValuesFromContext(r.Context())
}

// Handler returns an http.Handler that matches incoming request URLs against
// this template. If the request URL matches, the extracted variable values
// are stored in the request context and the inner handler h is called.
// If the URL does not match, the handler responds with 404 Not Found.
//
// The matching is performed against the full request URL path and query string
// (i.e., r.URL.RequestURI()). For templates that include a scheme and host,
// matching uses r.URL.String() instead.
//
// Extracted values are available in the inner handler via ValuesFromRequest
// or ValuesFromContext.
//
// Example usage with http.ServeMux:
//
//	mux := http.NewServeMux()
//	tmpl := uritemplate.MustParse("/users/{user}/repos{?sort,page}")
//	mux.Handle("/users/", tmpl.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//	    vals := uritemplate.ValuesFromRequest(r)
//	    user := vals["user"] // String value
//	    // ...
//	})))
func (t *Template) Handler(h http.Handler) http.Handler {
	return &templateHandler{
		template: t,
		handler:  h,
	}
}

// HandlerFunc returns an http.Handler that matches incoming request URLs
// against this template. It is a convenience wrapper around Handler that
// accepts a function instead of an http.Handler.
func (t *Template) HandlerFunc(f func(http.ResponseWriter, *http.Request)) http.Handler {
	return t.Handler(http.HandlerFunc(f))
}

// templateHandler implements http.Handler using a URI template for matching.
type templateHandler struct {
	template *Template
	handler  http.Handler
}

// ServeHTTP matches the request URL against the template. On match, it
// stores extracted values in the request context and calls the inner handler.
// On mismatch, it responds with 404 Not Found.
func (th *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	matchURI := r.URL.RequestURI()
	if th.templateHasScheme() {
		matchURI = r.URL.String()
	}

	vals, ok := th.template.Match(matchURI)
	if !ok {
		http.NotFound(w, r)
		return
	}

	ctx := context.WithValue(r.Context(), valuesKey, vals)
	th.handler.ServeHTTP(w, r.WithContext(ctx))
}

// templateHasScheme checks whether the template's first literal part
// contains "://" indicating a scheme+host template.
func (th *templateHandler) templateHasScheme() bool {
	for _, p := range th.template.parts {
		lit, ok := p.(literal)
		if !ok {
			return false
		}
		s := string(lit)
		for i := 0; i < len(s); i++ {
			if s[i] == ':' && i+2 < len(s) && s[i+1] == '/' && s[i+2] == '/' {
				return true
			}
			if s[i] == '/' {
				return false
			}
		}
		return false
	}
	return false
}
