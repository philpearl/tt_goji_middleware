package base

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCacheAWhile(t *testing.T) {

	tests := []struct {
		method        string
		expectExpires bool
		cacheControl  string
	}{
		{method: "GET", expectExpires: true, cacheControl: "public, max-age=3600"},
		{method: "POST", expectExpires: false},
		{method: "PATCH", expectExpires: false},
		{method: "PUT", expectExpires: false},
		{method: "DELETE", expectExpires: false},
		{method: "OPTIONS", expectExpires: false},
	}

	for _, test := range tests {
		c := makeEnv()
		w := httptest.NewRecorder()
		m := CacheAWhileMiddleWare(time.Hour)

		h := m(&c, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		r, _ := http.NewRequest(test.method, "/", nil)

		start := time.Now()
		h.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("expect 200, get %d, %s", w.Code, w.Body.String())
		}

		expires := w.HeaderMap.Get("Expires")
		if test.expectExpires {
			if expires == "" {
				t.Errorf("Expected Expires header")
			}
			exp, err := time.Parse(time.RFC1123, expires)
			if err != nil {
				t.Errorf("Failed to parse expiry header. %v (%s)", err, expires)
			}

			if exp.Sub(start) > time.Hour+time.Minute {
				t.Errorf("expiry too far in the future")
			}

			if exp.Sub(start) < time.Hour-time.Minute {
				t.Errorf("expiry not far enough in the future")
			}

			cc := w.HeaderMap.Get("Cache-Control")
			if cc != test.cacheControl {
				t.Errorf("expected cache control %s, got %s", test.cacheControl, cc)
			}
		} else {
			if expires != "" {
				t.Errorf("Expected no expires header")
			}
		}
	}
}
