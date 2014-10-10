package tt_goji_middleware

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/zenazn/goji/web"
)

type gzipResponseWriter struct {
	// The http response we are wrapping
	Wrapped http.ResponseWriter
	// Have we written a status code and header?
	headerWritten bool
	// a gzip writer (which wraps Wrapped)
	writer *gzip.Writer
	// http status code written
	status int
}

func newGzipResponseWriter(wrapped http.ResponseWriter) *gzipResponseWriter {
	return &gzipResponseWriter{
		Wrapped: wrapped,
		writer:  gzip.NewWriter(wrapped),
	}
}

func (w *gzipResponseWriter) Close() {
	w.writer.Close()
}

func (w *gzipResponseWriter) Header() http.Header {
	return w.Wrapped.Header()
}

func (w *gzipResponseWriter) Write(data []byte) (int, error) {
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

func (w *gzipResponseWriter) WriteHeader(status int) {
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

/*
Simple GZIP middleware

GZIPs GET 200 OK responses if the request Accept-Encoding includes gzip.  Sets Content-Encoding to "gzip".
*/
func GzipMiddleWare(c *web.C, h http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		var gw *gzipResponseWriter
		if acceptsGzip(r) {
			gw = newGzipResponseWriter(w)
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
