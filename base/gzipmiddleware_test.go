package base

import (
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGzipError(t *testing.T) {
	c := makeEnv()

	w := httptest.NewRecorder()

	h := GzipMiddleWare(&c, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad stuff", http.StatusBadRequest)
	}))

	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Set("Accept-Encoding", "gzip")

	h.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expect 400, get %d, %s", w.Code, w.Body.String())
	}

	if w.Body.String() != "bad stuff\n" {
		t.Errorf("body not as expected.  have \"%s\"", w.Body.String())
	}

	if w.HeaderMap.Get("Content-Type") != "text/plain; charset=utf-8" {
		t.Errorf("Content-Type not as expected - have %s", w.HeaderMap["Content-Type"])
	}

	if w.HeaderMap.Get("Content-Encoding") != "" {
		t.Errorf("expected no encoding - have %v", w.HeaderMap)
	}
}

func TestGzipRange(t *testing.T) {
	c := makeEnv()

	w := httptest.NewRecorder()

	h := GzipMiddleWare(&c, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Range", "bytes 2-4/11")
		w.WriteHeader(http.StatusPartialContent)
		w.Write([]byte("sup"))
	}))

	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Set("Accept-Encoding", "gzip")
	r.Header.Set("Range", "2-4")

	h.ServeHTTP(w, r)

	if w.Code != http.StatusPartialContent {
		t.Errorf("expect 206, get %d, %s", w.Code, w.Body.String())
	}

	if w.HeaderMap.Get("Content-Encoding") != "" {
		t.Errorf("Content should not be encoded - encoding is %s", w.HeaderMap.Get("Content-Encoding"))
	}

	if w.Body.String() != "sup" {
		t.Errorf("body not as expected.  have \"%s\"", w.Body.String())
	}

}

func TestGzipImage(t *testing.T) {
	tests := []struct {
		ct string
	}{
		{ct: "image/gif"},
		{ct: "image/jpeg"},
		{ct: "image/png"},
	}

	for _, test := range tests {
		c := makeEnv()

		w := httptest.NewRecorder()

		h := GzipMiddleWare(&c, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", test.ct)
			w.Write([]byte("super things"))
		}))

		r, _ := http.NewRequest("GET", "/", nil)
		r.Header.Set("Accept-Encoding", "gzip")

		h.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("expect 200, get %d, %s", w.Code, w.Body.String())
		}

		if w.HeaderMap.Get("Content-Encoding") != "" {
			t.Errorf("Content should not be encoded - encoding is %s", w.HeaderMap.Get("Content-Encoding"))
		}

		if w.Body.String() != "super things" {
			t.Errorf("body not as expected.  have \"%s\"", w.Body.String())
		}
	}
}

func TestGzipOK(t *testing.T) {
	c := makeEnv()

	w := httptest.NewRecorder()

	h := GzipMiddleWare(&c, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("super things"))
	}))

	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Set("Accept-Encoding", "gzip")

	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expect 200, get %d, %s", w.Code, w.Body.String())
	}

	gr, err := gzip.NewReader(w.Body)
	if err != nil {
		t.Errorf("couldn't create a gzip raader, %v", err)
	}
	body, err := ioutil.ReadAll(gr)
	if err != nil {
		t.Errorf("Couldn't read gzip content. %v", err)
	}

	if string(body) != "super things" {
		t.Errorf("body not as expected.  have \"%s\"", body)
	}

	if w.HeaderMap.Get("Content-Encoding") != "gzip" {
		t.Errorf("expected gzip encoding - have %v", w.HeaderMap)
	}
}
