package base

import (
	"errors"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/zenazn/goji/web"
)

var (
	ErrorSessionNotFound error = errors.New("No session found matching that Id")
)

/*
Session represents a web session
*/
type Session struct {
	Idd    string
	dirty  bool
	Values map[string]interface{}
}

/*
Get the id for this session
*/
func (s *Session) Id() string {
	return s.Idd
}

/*
Has anything been set in this Session since it was retrieved?
*/
func (s *Session) IsDirty() bool {
	return s.dirty
}

/*
Set whether this Session is dirty.  Normally this is done automatically
by Put
*/
func (s *Session) SetDirty(dirty bool) {
	s.dirty = dirty
}

/*
Retrieve a value from the Session
*/
func (s *Session) Get(key string) (interface{}, bool) {
	val, ok := s.Values[key]
	return val, ok
}

/*
Save a value in the Session
*/
func (s *Session) Put(key string, value interface{}) {
	s.Values[key] = value
	s.dirty = true
}

/*
SessionHolder is an interface for a session repository.
*/
type SessionHolder interface {
	/*
	   Get the session associated with the current request, if there is one.
	*/
	Get(c web.C, r *http.Request) (*Session, error)

	/*
	   Create a new session

	   Note the implementation should create the session dirty so it will be saved
	   when the response is processed.  It should also add the session to c.Env["session"]
	   for easy access
	*/
	Create(c web.C) *Session

	/*
	   Destroy a session so that it can no longer be retrieved
	*/
	Destroy(c web.C, session *Session) error

	/*
	   Save a session
	*/
	Save(c web.C, session *Session) error

	/*
		AddToResponse writes the session cookie into the http response
	*/
	AddToResponse(c web.C, session *Session, w http.ResponseWriter)
}

/*
BaseSessionHolder is a building block you can use to build a SessionHolder implementation
*/
type BaseSessionHolder struct {
	Timeout int
}

/*
Create builds a session object, adds it to c.Env["session"] and marks it as dirty
*/
func (sh *BaseSessionHolder) Create(c web.C) *Session {
	session := &Session{
		Idd:    generateSessionId(),
		Values: make(map[string]interface{}, 0),
		dirty:  true,
	}
	c.Env["session"] = session
	return session
}

/*
GetSessionId extracts the session ID from a request if one is present.

Note this is not part of the SessionHolder interface - it is intended to
be used as a building block by SessionHolder implementations
*/
func (hs *BaseSessionHolder) GetSessionId(r *http.Request) string {
	cookie, err := r.Cookie("sessionid")
	if err != nil {
		return ""
	}
	return cookie.Value
}

func generateSessionId() string {
	a := uint64(rand.Int63())
	b := uint64(rand.Int63())

	return strconv.FormatUint(a, 36) + strconv.FormatUint(b, 36)
}

/*
Add the Session to the response

Basically this means setting a cookie
*/
func (sh *BaseSessionHolder) AddToResponse(c web.C, session *Session, w http.ResponseWriter) {
	cookie := http.Cookie{
		Name:   "sessionid",
		Value:  session.Id(),
		Path:   "/",
		MaxAge: sh.Timeout,
	}
	http.SetCookie(w, &cookie)
}

/*
MemorySessionHolder is an in-memory session holder.  Only really useful for testing
*/
type MemorySessionHolder struct {
	BaseSessionHolder
	store map[string]*Session
}

func NewMemorySessionHolder(timeout int) SessionHolder {
	return &MemorySessionHolder{
		BaseSessionHolder: BaseSessionHolder{
			Timeout: timeout,
		},
		store: make(map[string]*Session, 0),
	}
}

func (sh *MemorySessionHolder) Get(c web.C, r *http.Request) (*Session, error) {
	sessionId := sh.GetSessionId(r)
	session, ok := sh.store[sessionId]
	if !ok {
		return nil, ErrorSessionNotFound
	}
	return session, nil
}

func (sh *MemorySessionHolder) Destroy(c web.C, session *Session) error {
	delete(sh.store, session.Id())
	return nil
}

/*
   Save a session
*/
func (sh *MemorySessionHolder) Save(c web.C, session *Session) error {
	sh.store[session.Id()] = session
	return nil
}

/*
BuildSessionMiddleware builds middleware with the provided SessionHolder.  The middleware

 - adds the SessionHolder to c.Env["sessionholder"] so application code can create and delete sessions
 - finds the session associated with the request (if any) and puts it in c.Env["session"]
 - saves the session if dirty and the request is processed without a panic

Add the middleware as follows

  mux := web.New()
  mux.Use(redis.BuildRedis(":6379"))
  sh := redis.NewSessionHolder()
  mux.Use(redis.BuildSessionMiddleware(sh))
  // Add handlers that use sessions

The middleware does not create new sessions.  To create a session do the following.

  session := c.Env["session"].(*Session)
  if session != nil {
	  sh := c.Env["sessionholder"].(SessionHolder)
	  session := sh.Create()
  }
*/
func BuildSessionMiddleware(sh SessionHolder) func(c *web.C, h http.Handler) http.Handler {
	return func(c *web.C, h http.Handler) http.Handler {
		// Always store the session holder in context so that handlers can create or destroy sessions
		c.Env["sessionholder"] = sh

		handler := func(w http.ResponseWriter, r *http.Request) {
			session, err := sh.Get(*c, r)
			if err == nil {
				session.SetDirty(false)
				c.Env["session"] = session
			}
			h.ServeHTTP(w, r)

			// TODO: should this be saved via defer() so it even happens after a panic
			// TODO: retrieve from response?
			session = c.Env["session"].(*Session)
			if session != nil && session.IsDirty() {
				sh.Save(*c, session)
			}
		}
		return http.HandlerFunc(handler)
	}
}
