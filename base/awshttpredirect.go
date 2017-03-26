package base

import (
	"net/http"

	"github.com/zenazn/goji/web"
)

// AWSHTTPRedirect redirects traffic on port 80 to https. It does this by
// looking at the X-Forwarded-Proto header, which is set by AWS load balancers
func AWSHTTPRedirect(c *web.C, h http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		proto := r.Header.Get("X-Forwarded-Proto")
		if proto == "http" {
			r.URL.Scheme = "https"
			http.Redirect(w, r, r.URL.String(), http.StatusMovedPermanently)
		} else {
			h.ServeHTTP(w, r)
		}
	}
	return http.HandlerFunc(handler)
}
