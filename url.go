package uritemplate

import (
	"fmt"
	"net/url"
)

// MustParse is like Parse but panics on error. Useful for package-level template constants.
func MustParse(template string) *Template {
	t, err := Parse(template)
	if err != nil {
		panic(fmt.Sprintf("uritemplate: %s", err))
	}
	return t
}

// URL expands the template with the given values and parses the result as a *url.URL.
// Returns an error if expansion fails or the result is not a valid URL.
func (t *Template) URL(vars Values) (*url.URL, error) {
	s, err := t.Expand(vars)
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(s)
	if err != nil {
		return nil, fmt.Errorf("uritemplate: expanded template is not a valid URL: %w", err)
	}
	return u, nil
}

// FromURL extracts template variable values by matching the given URL against
// this template. Returns the extracted values and true if the URL matches,
// or nil and false otherwise.
//
// The URL is converted back to its string representation and matched against
// the template. Query parameters are matched by name for query-style operators
// ({?vars} and {&vars}), making the match independent of parameter ordering.
func (t *Template) FromURL(u *url.URL) (Values, bool) {
	vals, ok := t.Match(u.String())
	if !ok {
		return nil, false
	}

	// For query-style operators (? and &), use u.Query() to extract values
	// by name, ensuring order independence.
	query := u.Query()
	for _, p := range t.parts {
		expr, ok := p.(*expression)
		if !ok {
			continue
		}
		if expr.operator.prefix != "?" && expr.operator.prefix != "&" {
			continue
		}
		for _, vs := range expr.vars {
			qvals, exists := query[vs.name]
			if !exists || len(qvals) == 0 {
				delete(vals, vs.name)
				continue
			}
			if len(qvals) == 1 {
				vals[vs.name] = String(qvals[0])
			} else {
				vals[vs.name] = List(qvals...)
			}
		}
	}

	return vals, true
}
