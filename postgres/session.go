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
		content bytea
	)`
)

/*
SessionHolder is a Postgres-backed session store. It uses the postgres bytea data type and
the gob encoding to store the session

It requires an sql connection set up in c.Env["db"].
*/
type SessionHolder struct {
	base.BaseSessionHolder
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

	db := c.Env["db"].(*sql.DB)
	var session base.Session
	values := sessionValues{}

	err := db.QueryRow("SELECT content FROM sessions WHERE id=$1", sessionId).Scan(values)
	if err == nil {
		session.Values = values
		session.SetId(sessionId)
	}

	return &session, err
}

/*
Destroy deletes a session from postgres
*/
func (sh *SessionHolder) Destroy(c web.C, session *base.Session) error {
	sessionId := session.Id()
	delete(c.Env, "session")
	db := c.Env["db"].(*sql.DB)

	_, err := db.Exec("DELETE FROM sessions WHERE id=$1", sessionId)
	return err
}

/*
Save a session to postgres
*/
func (sh *SessionHolder) Save(c web.C, session *base.Session) error {
	sessionId := session.Id()
	db := c.Env["db"].(*sql.DB)

	return runTransaction(db, func(tx *sql.Tx) {
		_, err := db.Exec("INSERT INTO sessions (id, content) VALUES ($1, $2)", sessionId, sessionValues(session.Values))
		if err != nil && isAlreadyExists(err, "sessions") {
			_, err = db.Exec("UPDATE sessions SET content=$2 WHERE id=$1", sessionId, sessionValues(session.Values))
		}
		if err != nil {
			panic(err)
		}
	})
}

func (sh *SessionHolder) RegenerateId(c web.C, session *base.Session) (string, error) {
	sessionId := session.Id()
	newSessionId := sh.GenerateSessionId()
	db := c.Env["db"].(*sql.DB)

	_, err := db.Exec("UPDATE sessions SET id=$2 WHERE id=$1", sessionId, newSessionId)

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
	return nil
}
