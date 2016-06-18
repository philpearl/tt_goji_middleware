package base

import (
	"encoding/json"
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

type jsonLog struct {
	RemoteAddr     string `json:"remote_addr"`
	Method         string `json:"method"`
	RequestURI     string `json:"url"`
	Status         int    `json:"status"`
	ResponseTimeMs int64  `json:"response_ms"`
}

func LoggingMiddleWareJSON(c *web.C, h http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &StatusTrackingResponseWriter{w, http.StatusOK}
		h.ServeHTTP(ww, r)

		l := jsonLog{}
		fwd := r.Header.Get("X-Forwarded-For")
		if fwd == "" {
			l.RemoteAddr = r.RemoteAddr
		} else {
			l.RemoteAddr = fwd + ":" + r.Header.Get("X-Forwarded-Port")
		}
		l.Method = r.Method
		l.RequestURI = r.RequestURI
		l.Status = ww.Status
		l.ResponseTimeMs = time.Since(start).Nanoseconds() / 1000000

		data, err := json.Marshal(&l)
		if err != nil {
			log.Printf("Failed to marshal JSON log. %v", err)
		} else {
			log.Println(string(data))
		}
	}
	return http.HandlerFunc(handler)
}
