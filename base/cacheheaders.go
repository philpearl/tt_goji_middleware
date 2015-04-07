package base

import (
	"fmt"
	"net/http"
	"time"

	"github.com/zenazn/goji/web"
)

// CacheAWhileMiddleware builds a middleware that sets the Expires header for cacheFor into the future
func CacheAWhileMiddleWare(cacheFor time.Duration) func(c *web.C, h http.Handler) http.Handler {
	m := func(c *web.C, h http.Handler) http.Handler {
		handler := func(w http.ResponseWriter, r *http.Request) {

			if r.Method == "GET" {
				hdr := w.Header()
				var expires = time.Now().Add(cacheFor).Format(time.RFC1123)
				hdr.Set("Expires", expires)
				hdr.Set("Cache-Control", fmt.Sprintf("public, max-age=%d", cacheFor/time.Second))
			}

			h.ServeHTTP(w, r)
		}
		return http.HandlerFunc(handler)
	}
	return m
}
