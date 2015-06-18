package base

import (
	"net/http"

	"github.com/zenazn/goji/web"
)

type srResponseWriter struct {
	// The http response we are wrapping
	http.ResponseWriter
	// Have we written a status code and header?
	headerWritten bool
}

func newSrResponseWriter(wrapped http.ResponseWriter) *srResponseWriter {
	return &srResponseWriter{
		ResponseWriter: wrapped,
	}
}

func (w *srResponseWriter) stripAcceptRanges() {
	if !w.headerWritten {
		w.Header().Del("Accept-Ranges")
		w.headerWritten = true
	}
}

func (w *srResponseWriter) WriteHeader(status int) {
	w.stripAcceptRanges()
	w.ResponseWriter.WriteHeader(status)
}

func (w *srResponseWriter) Write(data []byte) (int, error) {
	w.stripAcceptRanges()
	return w.ResponseWriter.Write(data)
}

func StripRangeMiddleWare(c *web.C, h http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Strip any range request
		r.Header.Del("Range")
		// Don't let them accept ranges
		h.ServeHTTP(newSrResponseWriter(w), r)
	}
	return http.HandlerFunc(handler)
}
