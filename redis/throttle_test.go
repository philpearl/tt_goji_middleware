package redis

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	redigo "github.com/garyburd/redigo/redis"
	"github.com/zenazn/goji/web"
)

func TestThrottle(t *testing.T) {
	var err error
	c := &web.C{}
	c.Env = make(map[string]interface{}, 0)
	c.Env["redis"], err = redigo.Dial("tcp", ":6379")
	if err != nil {
		t.Fatalf("could not connect to redis")
	}

	m := BuildThrottleMiddleWare(10, func(c *web.C, r *http.Request) (string, int) {
		return c.Env["key"].(string), c.Env["limit"].(int)
	})

	r, err := http.NewRequest("GET", "http://example.com/foo", nil)
	if err != nil {
		t.Fatalf("couldn't create dummy request")
	}

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	start := int(time.Now().Unix())

	c.Env["key"] = "testthr"
	c.Env["limit"] = 10

	f := func(expLimit, expRem, expCode int) {
		w := httptest.NewRecorder()
		m(c, h).ServeHTTP(w, r)

		limit := getHeaderInt(w.HeaderMap, "X-RateLimit-Limit")
		if limit != expLimit {
			t.Fatalf("X-RateLimit-Limit expected %d was %d", expLimit, limit)
		}

		remaining := getHeaderInt(w.HeaderMap, "X-RateLimit-Remaining")
		if remaining != expRem {
			t.Fatalf("X-RateLimit-Remaining expected %d was %d", expRem, remaining)
		}

		reset := getHeaderInt(w.HeaderMap, "X-RateLimit-Reset")
		if reset < start+9 || reset > start+11 {
			t.Fatalf("reset a bit funny.  Value is %d, start is %d", reset, start)
		}

		if w.Code != expCode {
			t.Fatalf("unexpected status code %d", w.Code)
		}

	}

	f(10, 9, 200)
	f(10, 8, 200)
	f(10, 7, 200)
	f(10, 6, 200)
	f(10, 5, 200)
	f(10, 4, 200)
	f(10, 3, 200)
	f(10, 2, 200)
	f(10, 1, 200)
	f(10, 0, 200)
	f(10, -1, 429)
}

func getHeaderInt(h http.Header, key string) int {
	strval := h.Get(key)
	val, _ := strconv.Atoi(strval)
	return val
}
