/*
Package redis contains middleware that depends on github.com/garyburd/redigo/redis
*/
package redis

import (
	"net/http"
	"time"

	"github.com/zenazn/goji/web"

	redigo "github.com/garyburd/redigo/redis"
)

/*
A middleware that ensures a redis connection is present in c.Env["redis"].  The
connection is built from a redigo Pool.
*/
func BuildRedis(redisAddr string) func(c *web.C, h http.Handler) http.Handler {
	pool := &redigo.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redigo.Conn, error) {
			return redigo.Dial("tcp", redisAddr)
		},
	}

	return func(c *web.C, h http.Handler) http.Handler {
		// Establish a connection to redis and store it in the environment
		// Env["redis"].(redis.Conn)

		handler := func(w http.ResponseWriter, r *http.Request) {
			redis_conn := pool.Get()
			defer redis_conn.Close()
			c.Env["redis"] = redis_conn

			h.ServeHTTP(w, r)
		}
		return http.HandlerFunc(handler)
	}
}
