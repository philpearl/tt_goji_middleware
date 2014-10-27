package redis

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net/http"

	"github.com/philpearl/tt_goji_middleware/base"

	redigo "github.com/garyburd/redigo/redis"
	"github.com/zenazn/goji/web"
)

const (
	DEFAULT_SESSION_TIMEOUT = 30 * 24 * 60 * 60
)

func init() {
	// so we can encode the session properly
	gob.Register(map[string]interface{}{})
}

/*
SessionHolder is a redis-backed session store that gob-encodes sessions.

It requires a redigo redis connection set up in c.Env["redis"].  You can use
BuildRedis() to create middleware that does this
*/
type SessionHolder struct {
	base.BaseSessionHolder
}

/*
NewSessionHolder creates a new redis-backed gob-encoded session holder
*/
func NewSessionHolder() base.SessionHolder {
	return &SessionHolder{
		BaseSessionHolder: base.BaseSessionHolder{
			Timeout: DEFAULT_SESSION_TIMEOUT,
		},
	}
}

/*
Get the session for this request from Redis
*/
func (sh *SessionHolder) Get(c web.C, r *http.Request) (*base.Session, error) {
	sessionId := sh.GetSessionId(r)
	if sessionId == "" {
		return nil, base.ErrorSessionNotFound
	}

	conn := c.Env["redis"].(redigo.Conn)

	sess, err := conn.Do("GET", sessionKey(sessionId))
	if err != nil {
		return nil, err
	}
	if sess == nil {
		return nil, base.ErrorSessionNotFound
	}
	sessionBytes, err := redigo.Bytes(sess, err)
	if err != nil {
		return nil, err
	}

	dec := gob.NewDecoder(bytes.NewReader(sessionBytes))
	var session base.Session
	err = dec.Decode(&session)
	if err == nil {
		session.SetId(sessionId)
	}

	return &session, err
}

/*
Destroy deletes a session from redis
*/
func (sh *SessionHolder) Destroy(c web.C, session *base.Session) error {
	sessionId := session.Id()
	delete(c.Env, "session")
	conn := c.Env["redis"].(redigo.Conn)

	_, err := conn.Do("DEL", sessionKey(sessionId))
	return err
}

/*
Save a session to redis
*/
func (sh *SessionHolder) Save(c web.C, session *base.Session) error {
	sessionId := session.Id()
	conn := c.Env["redis"].(redigo.Conn)

	var b bytes.Buffer
	enc := gob.NewEncoder(&b)

	err := enc.Encode(session)
	if err != nil {
		return err
	}

	_, err = conn.Do("SET", sessionKey(sessionId), b.String(), "EX", sh.Timeout)

	return err
}

func sessionKey(sessionId string) string {
	return fmt.Sprintf("sess:%s", sessionId)
}
