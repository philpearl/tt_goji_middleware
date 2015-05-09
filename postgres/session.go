/* Package postgres contains code that depends on a Postgresql database backend, and github.com/lib/pq */
package postgres

import (
	"database/sql"
	"net/http"

	"github.com/philpearl/tt_goji_middleware/base"

	"github.com/zenazn/goji/web"
)

const (
	DEFAULT_SESSION_TIMEOUT = 30 * 24 * 60 * 60

	TABLE_DEFINITION = `CREATE TABLE IF NOT EXISTS sessions (
		id char(72) PRIMARY KEY,
		content bytea,
		expires timestamp with time zone
	)`
)

/*
SessionHolder is a Postgres-backed session store. It uses the postgres bytea data type and
the gob encoding to store the session
*/
type SessionHolder struct {
	base.BaseSessionHolder
	db *sql.DB
}

/*
NewSessionHolder creates a new postgres-backed session holder
*/
func NewSessionHolder(db *sql.DB) (base.SessionHolder, error) {
	_, err := db.Exec(TABLE_DEFINITION)
	if err != nil {
		return nil, err
	}

	return &SessionHolder{
		BaseSessionHolder: base.NewBaseSessionHolder(DEFAULT_SESSION_TIMEOUT),
		db:                db,
	}, nil
}

/*
Get the session for this request from Postgres
*/
func (sh *SessionHolder) Get(c web.C, r *http.Request) (*base.Session, error) {
	sessionId := sh.GetSessionId(r)
	if sessionId == "" {
		return nil, base.ErrorSessionNotFound
	}

	var session base.Session
	values := sessionValues{}

	err := sh.db.QueryRow("SELECT content FROM sessions WHERE id=$1", sessionId).Scan(values)
	if err == nil {
		session.Values = values
		session.SetId(sessionId)
	} else if err == sql.ErrNoRows {
		err = base.ErrorSessionNotFound
	}

	return &session, err
}

/*
Destroy deletes a session from postgres
*/
func (sh *SessionHolder) Destroy(c web.C, session *base.Session) error {
	sessionId := session.Id()
	delete(c.Env, "session")
	_, err := sh.db.Exec("DELETE FROM sessions WHERE id=$1", sessionId)
	return err
}

/*
Save a session to postgres
*/
func (sh *SessionHolder) Save(c web.C, session *base.Session) error {
	sessionId := session.Id()

	// There is a potential race here if the insert fails because the session exists, but it is deleted
	// before we can update it.
	_, err := sh.db.Exec("INSERT INTO sessions (id, content, expires) VALUES ($1, $2, now() + $3 * interval '1 second')", sessionId, sessionValues(session.Values), sh.Timeout)
	if err != nil && isAlreadyExists(err, "sessions") {
		_, err = sh.db.Exec("UPDATE sessions SET content=$2, expires=now() + $3 * interval '1 second' WHERE id=$1", sessionId, sessionValues(session.Values), sh.Timeout)
	}
	return err
}

func (sh *SessionHolder) RegenerateId(c web.C, session *base.Session) (string, error) {
	sessionId := session.Id()
	newSessionId := sh.GenerateSessionId()

	_, err := sh.db.Exec("UPDATE sessions SET id=$2 WHERE id=$1", sessionId, newSessionId)

	if err == nil {
		// This all worked, use the new session Id
		session.SetId(newSessionId)
	} else {
		// Move back to the old session Id.  We hope this is a transitory error...
		newSessionId = sessionId
	}
	return newSessionId, err
}

func (sh *SessionHolder) ResetTTL(c web.C, session *base.Session) error {
	// Need to implement a TTL...
	_, err := sh.db.Exec("UPDATE sessions SET expires=now()+$2 * interval '1 second' WHERE id=$1", session.Id(), sh.Timeout)
	return err
}
