package base

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStripRange(t *testing.T) {
	c := makeEnv()

	w := httptest.NewRecorder()

	h := StripRangeMiddleWare(&c, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Range") != "" {
			t.Fatalf("request has range header %s", r.Header.Get("Range"))
		}
		w.Header().Set("Accept-Ranges", "bytes")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello"))
	}))

	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Set("Range", "2-4")

	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expect 206, get %d, %s", w.Code, w.Body.String())
	}

	if w.HeaderMap.Get("Accept-Ranges") != "" {
		t.Errorf("Should not have accept ranges")
	}

	if w.Body.String() != "hello" {
		t.Errorf("body not as expected.  have \"%s\"", w.Body.String())
	}
}

func TestStripRangeNoWriteHeader(t *testing.T) {
	c := makeEnv()

	w := httptest.NewRecorder()

	h := StripRangeMiddleWare(&c, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Range") != "" {
			t.Fatalf("request has range header %s", r.Header.Get("Range"))
		}
		w.Header().Set("Accept-Ranges", "bytes")
		w.Write([]byte("hello"))
	}))

	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Set("Range", "2-4")

	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expect 200, get %d, %s", w.Code, w.Body.String())
	}

	if w.HeaderMap.Get("Accept-Ranges") != "" {
		t.Errorf("Should not have accept ranges")
	}

	if w.Body.String() != "hello" {
		t.Errorf("body not as expected.  have \"%s\"", w.Body.String())
	}
}
