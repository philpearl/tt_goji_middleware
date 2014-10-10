package tt_goji_middleware

import (
	"net/http"

	"github.com/zenazn/goji/web"
)

// Build a middleware that always sets c.Env[key] = value
func BuildEnvSet(key string, value interface{}) func(c *web.C, h http.Handler) http.Handler {
	return func(c *web.C, h http.Handler) http.Handler {
		handler := func(w http.ResponseWriter, r *http.Request) {
			c.Env[key] = value
			h.ServeHTTP(w, r)
		}
		return http.HandlerFunc(handler)
	}
}

// Build a middleware that always sets c.Env[key] = *value
func BuildUpdateableSet(key string, value *interface{}) func(c *web.C, h http.Handler) http.Handler {
	return func(c *web.C, h http.Handler) http.Handler {
		handler := func(w http.ResponseWriter, r *http.Request) {
			c.Env[key] = *value
			h.ServeHTTP(w, r)
		}
		return http.HandlerFunc(handler)
	}
}
