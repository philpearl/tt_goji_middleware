package base

import (
	crand "crypto/rand"
	"errors"
	"log"
	"math/big"
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
	id     string
	dirty  bool
	Values map[string]interface{}
}

/*
Get the id for this session
*/
func (s *Session) Id() string {
	return s.id
}

func (s *Session) SetId(id string) {
	s.id = id
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
Delete a value from the session
*/
func (s *Session) Del(key string) {
	delete(s.Values, key)
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
	   Regenerate the sessionId for an existing session. This can be used against session fixation attacks
	*/
	RegenerateId(c web.C, session *Session) (string, error)
	
	/*
		AddToResponse writes the session cookie into the http response
	*/
	AddToResponse(c web.C, session *Session, w http.ResponseWriter)
	
	/* SetTimeout sets the timeout for session data */
	SetTimeout(timeout int)
	
	/* GetTimeout retrieves the currently set timeout for session data */
	GetTimeout() int
	
	/* Set HttpOnly cookie on/off */
	SetHttpOnly(httpOnly bool)
	
	/* Set Secure cookie on/off */
	SetSecure(secure bool)
	
	/* ResetTTL can be implemented to reset the TTL for a session object if not dirty */
	ResetTTL(c web.C, session *Session) error
}

/*
BaseSessionHolder is a building block you can use to build a SessionHolder implementation
*/
type BaseSessionHolder struct {
	Timeout    int
	RandSource rand.Source
	HttpOnly   bool
	Secure     bool
}

func NewBaseSessionHolder(timeout int) BaseSessionHolder {
	seed, err := crand.Int(crand.Reader, big.NewInt(0x7FFFFFFFFFFFFFFF))
	if err != nil {
		log.Panicf("Could not get a random seed for session generation, %v", err)
	}

	return BaseSessionHolder{
		Timeout:    timeout,
		RandSource: rand.NewSource(seed.Int64()),
		HttpOnly:   false,
		Secure:     false,
	}
}

/*
Create builds a session object, adds it to c.Env["session"] and marks it as dirty
*/
func (sh *BaseSessionHolder) Create(c web.C) *Session {
	session := &Session{
		id:     sh.GenerateSessionId(),
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
func (sh *BaseSessionHolder) GetSessionId(r *http.Request) string {
	cookie, err := r.Cookie("sessionid")
	if err != nil {
		return ""
	}
	return cookie.Value
}

/*
GenerateSessionId generates an id that can be used as sessionid

Note this is not part of the SessionHolder interface - it is intended to
be used as a building block by SessionHolder implementations
*/

func (sh *BaseSessionHolder) GenerateSessionId() string {
	a := uint64(sh.RandSource.Int63())
	b := uint64(sh.RandSource.Int63())

	return strconv.FormatUint(a, 36) + strconv.FormatUint(b, 36)
}

/*
GetTimeout retrieves the currently set TTL for session objects 
*/
func (sh *BaseSessionHolder) GetTimeout() int {
	return sh.Timeout
}

/*
SetTimeout updates the TTL for session objects

Note that the TTL will only be updated for existing sessions when the session is requested again, it might timeout if
this request takes too long to occur
*/
func (sh *BaseSessionHolder) SetTimeout(timeout int) {
	sh.Timeout = timeout
}

func (sh *BaseSessionHolder) SetHttpOnly(httpOnly bool) {
	sh.HttpOnly = httpOnly
}

func (sh *BaseSessionHolder) SetSecure(secure bool) {
	sh.Secure = secure
}

/*
Add the Session to the response

Basically this means setting a cookie
*/
func (sh *BaseSessionHolder) AddToResponse(c web.C, session *Session, w http.ResponseWriter) {
	cookie := http.Cookie{
		Name:     "sessionid",
		Value:    session.Id(),
		Path:     "/",
		HttpOnly: sh.HttpOnly,
		Secure:   sh.Secure,
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
		BaseSessionHolder: NewBaseSessionHolder(timeout),
		store:             make(map[string]*Session, 0),
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
	delete(c.Env, "session")
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
   Regenerate the session id of an existing session.
*/
func (sh *MemorySessionHolder) RegenerateId(_ web.C, session *Session) (string, error) {
	newSessionId := sh.GenerateSessionId()
	delete(sh.store, session.Id())
	session.SetId(newSessionId)
	sh.store[newSessionId] = session
	return newSessionId, nil
}

func (sh *MemorySessionHolder) ResetTTL(_ web.C, session *Session) error {
	return nil
}

func SessionFromEnv(c *web.C) (*Session, bool) {
	s, ok := c.Env["session"]
	if ok && s != nil {
		return s.(*Session), true
	}
	return nil, false
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
*/
func BuildSessionMiddleware(sh SessionHolder) func(c *web.C, h http.Handler) http.Handler {
	return func(c *web.C, h http.Handler) http.Handler {
		handler := func(w http.ResponseWriter, r *http.Request) {
			// Always store the session holder in context so that handlers can create or destroy sessions
			c.Env["sessionholder"] = sh

			session, err := sh.Get(*c, r)
			if err == nil {
				session.SetDirty(false)
				c.Env["session"] = session
			} else {
				if err == ErrorSessionNotFound {
					session = sh.Create(*c)
					// add sessionid to cookie
					sh.AddToResponse(*c, session, w)
				} else {
					log.Printf("error loading session %v", err)
					http.Error(w, "Failed to load session data", http.StatusServiceUnavailable)
					return
				}
			}
			h.ServeHTTP(w, r)

			// TODO: should this be saved via defer() so it even happens after a panic
			// TODO: retrieve from response?
			session, ok := SessionFromEnv(c)
			if ok {
				if session.IsDirty() {
					err := sh.Save(*c, session)
					if err != nil {
						log.Printf("Failed to save session - %v", err)
					}
				} else {
					/* not dirty but our sessionhandler might need to update the timeout/expiration */
					err := sh.ResetTTL(*c, session)
					if err != nil {
						log.Printf("Failed to update TTL for session - %v", err)
					}
				}
			}
		}
		return http.HandlerFunc(handler)
	}
}
