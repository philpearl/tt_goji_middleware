package tt_goji_middleware

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	// "strings"
	"runtime"

	"github.com/zenazn/goji/web"

	"github.com/kisielk/raven-go/raven"
)

func BuildErrorCatcher(sentryDSN string) func(c *web.C, h http.Handler) http.Handler {
	var sentryClient *raven.Client
	if sentryDSN != "" {
		var err error
		sentryClient, err = raven.NewClient(sentryDSN)
		if err != nil {
			log.Printf("Couldn't connect to sentry %v\n", err)
		}
	}

	return func(c *web.C, h http.Handler) http.Handler {
		handler := func(w http.ResponseWriter, r *http.Request) {

			defer func() {
				err := recover()
				if err == nil {
					return
				}
				if sentryClient != nil {
					// Send the error to sentry
					const size = 1 << 12
					buf := make([]byte, size)
					n := runtime.Stack(buf, false)
					sentryClient.CaptureMessage(fmt.Sprintf("%v\nStacktrace:\n%s", err, buf[:n]))
				}

				switch err := err.(type) {
				case HttpError:
					err.WriteResponse(w)
					return
				default:
					http.Error(w, http.StatusText(500), 500)
					log.Printf("Panic: %v\n", err)
					debug.PrintStack()
					return
				}
			}()

			h.ServeHTTP(w, r)
		}
		return http.HandlerFunc(handler)
	}
}

type HttpError struct {
	StatusCode int
	Message    string
}

func (h HttpError) Error() string {
	return h.Message
}

func (h HttpError) WriteResponse(w http.ResponseWriter) {
	http.Error(w, h.Message, h.StatusCode)
}

func MakeError(statusCode int, format string, params ...interface{}) error {
	return HttpError{statusCode, fmt.Sprintf(format, params...)}
}

func ThrowError(statusCode int, format string, params ...interface{}) {
	panic(MakeError(statusCode, format, params...))
}
