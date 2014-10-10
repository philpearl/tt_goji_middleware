package tt_goji_middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/zenazn/goji/web"
)

/*
Middleware that logs responses.

The output format is:

<remote addr> - <method> <url> <status code> <response time ms>

Remote address is taken from X-Forwarded-For & X-Forwarded-Port if present
*/
func LoggingMiddleWare(c *web.C, h http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &StatusTrackingResponseWriter{w, http.StatusOK}
		h.ServeHTTP(ww, r)

		var remoteAddr string
		fwd := r.Header.Get("X-Forwarded-For")
		if fwd == "" {
			remoteAddr = r.RemoteAddr
		} else {
			remoteAddr = fwd + ":" + r.Header.Get("X-Forwarded-Port")
		}
		log.Printf("%s - %s %s %d %dms\n", remoteAddr, r.Method, r.RequestURI, ww.Status, time.Since(start).Nanoseconds()/1000000)
	}
	return http.HandlerFunc(handler)
}
