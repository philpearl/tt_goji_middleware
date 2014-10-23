package base

import (
	"net/http"
	"regexp"

	"github.com/zenazn/goji/web"
)

/*
BuildStripPrefix builds Goji middleware that strips a prefix from the request URL path.

pattern can be either a string or a *regexp.Regexp.  If it is a string a prefix of the same length
as the string is removed from the path.  If it is a Regexp then everything up to the end of the first
match is removed
*/
func BuildStripPrefix(pattern interface{}) func(c *web.C, h http.Handler) http.Handler {
	return func(c *web.C, h http.Handler) http.Handler {
		handler := func(w http.ResponseWriter, r *http.Request) {
			// Now alter the request to remove this first part of the path.  This
			// is infuriating as the Path is stored decoded, so you can't really know
			// which / are really path separators
			switch pattern := pattern.(type) {
			case *regexp.Regexp:
				loc := pattern.FindStringIndex(r.URL.Path)
				r.URL.Path = r.URL.Path[loc[1]-1 : len(r.URL.Path)]
			case string:
				r.URL.Path = r.URL.Path[len(pattern):]
			}
			h.ServeHTTP(w, r)
		}
		return http.HandlerFunc(handler)
	}
}
