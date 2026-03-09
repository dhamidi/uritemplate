package uritemplate

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_Match(t *testing.T) {
	tmpl := MustParse("/users/{user}")
	var got Values
	h := tmpl.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = ValuesFromRequest(r)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/users/alice", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if got == nil {
		t.Fatal("expected values, got nil")
	}
	if v, ok := got["user"]; !ok || v.str != "alice" {
		t.Fatalf("expected user=alice, got %v", got)
	}
}

func TestHandler_NoMatch(t *testing.T) {
	tmpl := MustParse("/users/{user}/repos")
	h := tmpl.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called on mismatch")
	})

	req := httptest.NewRequest("GET", "/posts/123", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_QueryParams(t *testing.T) {
	tmpl := MustParse("/search{?q,lang}")
	var got Values
	h := tmpl.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = ValuesFromRequest(r)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/search?q=hello&lang=en", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if got == nil {
		t.Fatal("expected values, got nil")
	}
	if v := got["q"]; v.str != "hello" {
		t.Fatalf("expected q=hello, got %v", v)
	}
	if v := got["lang"]; v.str != "en" {
		t.Fatalf("expected lang=en, got %v", v)
	}
}

func TestHandler_PathAndQuery(t *testing.T) {
	tmpl := MustParse("/users/{user}/repos{?sort}")
	var got Values
	h := tmpl.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = ValuesFromRequest(r)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/users/alice/repos?sort=stars", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if got == nil {
		t.Fatal("expected values, got nil")
	}
	if v := got["user"]; v.str != "alice" {
		t.Fatalf("expected user=alice, got %v", v)
	}
	if v := got["sort"]; v.str != "stars" {
		t.Fatalf("expected sort=stars, got %v", v)
	}
}

func TestHandler_WithServeMux(t *testing.T) {
	tmpl := MustParse("/users/{user}")
	mux := http.NewServeMux()
	mux.Handle("/users/", tmpl.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vals := ValuesFromRequest(r)
		fmt.Fprintf(w, "user=%s", vals["user"].str)
	}))

	req := httptest.NewRequest("GET", "/users/bob", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if body := rec.Body.String(); body != "user=bob" {
		t.Fatalf("expected body 'user=bob', got %q", body)
	}
}

func TestHandler_HandlerFunc(t *testing.T) {
	tmpl := MustParse("/items/{id}")
	called := false

	// Using Handler with http.HandlerFunc
	h1 := tmpl.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		vals := ValuesFromRequest(r)
		if v := vals["id"]; v.str != "42" {
			t.Fatalf("expected id=42, got %v", v)
		}
	}))

	req := httptest.NewRequest("GET", "/items/42", nil)
	rec := httptest.NewRecorder()
	h1.ServeHTTP(rec, req)
	if !called {
		t.Fatal("Handler was not called")
	}

	// Using HandlerFunc directly
	called = false
	h2 := tmpl.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		vals := ValuesFromRequest(r)
		if v := vals["id"]; v.str != "42" {
			t.Fatalf("expected id=42, got %v", v)
		}
	})

	req = httptest.NewRequest("GET", "/items/42", nil)
	rec = httptest.NewRecorder()
	h2.ServeHTTP(rec, req)
	if !called {
		t.Fatal("HandlerFunc was not called")
	}
}

func TestValuesFromContext_NoValues(t *testing.T) {
	ctx := context.Background()
	vals := ValuesFromContext(ctx)
	if vals != nil {
		t.Fatalf("expected nil, got %v", vals)
	}
}

func TestHandler_SchemeTemplate(t *testing.T) {
	tmpl := MustParse("https://api.example.com/users/{user}")
	var got Values
	h := tmpl.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = ValuesFromRequest(r)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "https://api.example.com/users/charlie", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if got == nil {
		t.Fatal("expected values, got nil")
	}
	if v := got["user"]; v.str != "charlie" {
		t.Fatalf("expected user=charlie, got %v", v)
	}
}
