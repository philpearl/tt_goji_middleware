package tt_goji_middleware

import (
	"net/http"

	"github.com/zenazn/goji/web"
)

func BuildEnvSet(key string, value interface{}) func(c *web.C, h http.Handler) http.Handler {
	// Build a middleware that always set c.Env[key] = value
	return func(c *web.C, h http.Handler) http.Handler {
		handler := func(w http.ResponseWriter, r *http.Request) {
			c.Env[key] = value
			h.ServeHTTP(w, r)
		}
		return http.HandlerFunc(handler)
	}
}

func BuildUpdateableSet(key string, value *interface{}) func(c *web.C, h http.Handler) http.Handler {
	// Build a middleware that always set c.Env[key] = value
	return func(c *web.C, h http.Handler) http.Handler {
		handler := func(w http.ResponseWriter, r *http.Request) {
			c.Env[key] = *value
			h.ServeHTTP(w, r)
		}
		return http.HandlerFunc(handler)
	}
}
