package base

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHttpRedirect(t *testing.T) {
	c := makeEnv()

	w := httptest.NewRecorder()

	h := AWSHTTPRedirect(&c, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Shouldn't get here")
	}))

	r, _ := http.NewRequest("GET", "http://fred.com/hat", nil)
	r.Header.Set("X-Forwarded-Proto", "http")

	h.ServeHTTP(w, r)

	if w.Code != http.StatusMovedPermanently {
		t.Errorf("expect 301, get %d, %s", w.Code, w.Body.String())
	}

	loc := w.HeaderMap.Get("Location")
	if loc != "https://fred.com/hat" {
		t.Errorf("Redirect location not as expected. Have %s", loc)
	}
}

func TestHttpRedirectNoRedirect(t *testing.T) {
	c := makeEnv()

	w := httptest.NewRecorder()

	h := AWSHTTPRedirect(&c, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r, _ := http.NewRequest("GET", "http://fred.com/hat", nil)
	r.Header.Set("X-Forwarded-Proto", "https")

	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expect 200, get %d, %s", w.Code, w.Body.String())
	}

	loc := w.HeaderMap.Get("Location")
	if loc != "" {
		t.Errorf("Redirect location not as expected. Have %s", loc)
	}
}
