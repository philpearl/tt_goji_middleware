package tt_goji_middleware

import (
	"net/http"
)

type StatusTrackingResponseWriter struct {
	http.ResponseWriter
	// http status code written
	Status int
}

func (w *StatusTrackingResponseWriter) WriteHeader(status int) {
	w.Status = status
	w.ResponseWriter.WriteHeader(status)
}
