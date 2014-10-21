package redis

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	redigo "github.com/garyburd/redigo/redis"
	"github.com/zenazn/goji/web"
)

type Keyfunc func(c *web.C, r *http.Request) (string, int)

/*
Throttling middleware using redis.

Parameters

  interval - throttling interval in seconds
  keyfunc - function that looks at the current request and returns an appropriate throttle key and limit

Assumes redis connection is in c.Env["redis"] - see BuildRedis()


Example

  	m := web.New()
	m.Use(middleware.EnvInit)
	m.Use(redis.BuildRedis(config.RedisAddr))
	m.Use(IdentifyServiceMiddleware)
	m.Use(redis.BuildThrottleMiddleWare(3600, func(c *web.C, r *http.Request) (string, int) {
		service_id := c.Env["service_id"].(int)
		return fmt.Sprintf("api:throttle:%d", service_id), 1000
	}))

*/
func BuildThrottleMiddleWare(interval int, keyfunc Keyfunc) func(c *web.C, h http.Handler) http.Handler {

	// Script increments the key, and sets expiry to "interval" seconds if the value is 1
	var redisThrottleScript = redigo.NewScript(
		1,
		fmt.Sprintf(
			`local current; current = redis.call('incr',KEYS[1]);if tonumber(current) == 1 then redis.call('expire', KEYS[1], %d) end return {current, redis.call('ttl', KEYS[1])}`,
			interval,
		),
	)

	return func(c *web.C, h http.Handler) http.Handler {
		handler := func(w http.ResponseWriter, r *http.Request) {
			// Get a throttle key that identifies this flow
			throttleKey, limit := keyfunc(c, r)
			if limit > 0 {
				redis_conn := c.Env["redis"].(redigo.Conn)
				// Returns number of requests on this key in this period, and TTL of this period
				rsp, err := redigo.Values(redisThrottleScript.Do(redis_conn, throttleKey))
				if err != nil {
					log.Printf("Throttling: Cache failure, %v", err)
					http.Error(w, fmt.Sprintf("Throttling: Cache failure, %v", err), http.StatusServiceUnavailable)
				}
				numRequests := int(rsp[0].(int64))
				ttl := rsp[1].(int64)

				h := w.Header()
				setHeaderInt(h, "X-RateLimit-Limit", limit)
				setHeaderInt(h, "X-RateLimit-Remaining", limit-numRequests)
				setHeaderInt64(h, "X-Ratelimit-Reset", time.Now().Unix()+ttl)
				if numRequests > limit {
					http.Error(w, fmt.Sprintf("Request rate limit exceeded - allowed rate is %d requests every %d seconds", limit, interval), 429)
					return
				}
			}

			h.ServeHTTP(w, r)
		}
		return http.HandlerFunc(handler)
	}
}

func setHeaderInt(h http.Header, key string, val int) {
	setHeaderInt64(h, key, int64(val))
}

func setHeaderInt64(h http.Header, key string, val int64) {
	h.Set(key, strconv.FormatInt(val, 10))
}
