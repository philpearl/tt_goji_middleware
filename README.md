# Some Goji middleware

Some simple Goji middleware I've found useful.  Find the documentation at http://godoc.org/github.com/philpearl/tt_goji_middleware.

[![Build Status](https://travis-ci.org/philpearl/tt_goji_middleware.svg)](https://travis-ci.org/philpearl/tt_goji_middleware)

The middleware is arranged in packages based on external dependencies

- base has no external dependencies except Goji
- raven depends on github.com/kisielk/raven-go/raven
- redis depends on github.com/garyburd/redigo/redis

Just go get the sub-packages you need.

## What's included

In base:
- Set something in Context for all requests.  For example global configuration or a database connection pool
- Error catching and reporting
- Logging ('fraid I don't like the Goji version)
- A very simple GZIP

In raven:
- Catch panics, log them, send responses and report them to Sentry

In redis:
- Ensure there's a redis connection in c.Env["redis"].  Connections come from a pool and are not opened until used.
- A Redis based rate limiter that issues a single command to Redis per request.
