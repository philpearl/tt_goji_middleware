package tt_goji_middleware

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/zenazn/goji/web"
)

type GzipResponseWriter struct {
	// The http response we are wrapping
	Wrapped http.ResponseWriter
	// Have we written a status code and header?
	headerWritten bool
	// a gzip writer (which wraps Wrapped)
	writer *gzip.Writer
	// http status code written
	status int
}

func NewGzipResponseWriter(wrapped http.ResponseWriter) *GzipResponseWriter {
	return &GzipResponseWriter{
		Wrapped: wrapped,
		writer:  gzip.NewWriter(wrapped),
	}
}

func (w *GzipResponseWriter) Close() {
	w.writer.Close()
}

func (w *GzipResponseWriter) Header() http.Header {
	return w.Wrapped.Header()
}

func (w *GzipResponseWriter) Write(data []byte) (int, error) {
	// We can't read what's written to our Wrapped response, so we need to track things ourselves
	if !w.headerWritten {
		w.WriteHeader(http.StatusOK)
	}
	if w.status == http.StatusOK {
		// Only use our gzip wrapper for OK responses.
		return w.writer.Write(data)
	}
	return w.Wrapped.Write(data)
}

func (w *GzipResponseWriter) WriteHeader(status int) {
	if !w.headerWritten {
		w.headerWritten = true
		w.status = status
		if status == http.StatusOK {
			// If the status is OK we will use gzip, so let the far end know
			w.Wrapped.Header().Set("Content-Encoding", "gzip")
		}
	}
	w.Wrapped.WriteHeader(status)
}

func acceptsGzip(r *http.Request) bool {
	// Returns true if the request indicates support for GZIP
	if r.Method != "GET" {
		return false
	}
	accepts, ok := r.Header[http.CanonicalHeaderKey("Accept-Encoding")]
	if ok {
		for _, accept := range accepts {
			for _, word := range strings.Split(accept, ",") {
				if strings.TrimSpace(word) == "gzip" {
					return true
				}
			}
		}
	}
	return false
}

func GzipMiddleWare(c *web.C, h http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		var gw *GzipResponseWriter
		if acceptsGzip(r) {
			gw = NewGzipResponseWriter(w)
			w = gw
		}
		h.ServeHTTP(w, r)

		if gw != nil {
			// Only close here if we've completed this with no panic as the close writes to the underlying writer
			gw.Close()
		}
	}
	return http.HandlerFunc(handler)
}
