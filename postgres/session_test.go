package postgres

import (
	"testing"

	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/philpearl/tt_goji_middleware/base"
	"github.com/zenazn/goji/web"
)

func TestNoSession(t *testing.T) {
	db, err := sql.Open("postgres", "postgres://test:test@localhost:/test?sslmode=disable")
	if err != nil {
		t.Skipf("Cannot connect to postgres. %v", err)
	}
	c := web.C{
		Env: map[interface{}]interface{}{"db": db},
	}
	sh, err := NewSessionHolder(db)
	if err != nil {
		t.Fatalf("failed to create session holder - %v", err)
	}

	r, _ := http.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{
		Name:  "sessionid",
		Value: "zzpza",
	})

	_, err = sh.Get(c, r)
	if err != base.ErrorSessionNotFound {
		t.Fatalf("got error reading session from request, %v", err)
	}
}

func TestSessionUpdate(t *testing.T) {
	db, err := sql.Open("postgres", "postgres://test:test@localhost:/test?sslmode=disable")
	if err != nil {
		t.Skipf("Cannot connect to postgres. %v", err)
	}
	c := web.C{
		Env: map[interface{}]interface{}{"db": db},
	}
	sh, err := NewSessionHolder(db)
	if err != nil {
		t.Fatalf("failed to create session holder - %v", err)
	}

	// Create the session
	s := sh.Create(c)

	s.Put("cheese", "cheddar")
	err = sh.Save(c, s)
	if err != nil {
		t.Fatalf("failed to save session - %v", err)
	}

	s.Put("hat", 3)

	err = sh.Save(c, s)
	if err != nil {
		t.Fatalf("failed to save session on second attempt - %v", err)
	}

	r, _ := http.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{
		Name:  "sessionid",
		Value: s.Id(),
	})

	s1, err := sh.Get(c, r)
	if err != nil {
		t.Fatalf("got error reading session from request, %v", err)
	}

	val, ok := s1.Get("cheese")
	cheese := val.(string)
	if !ok || cheese != "cheddar" {
		t.Fatalf("could not get right cheese %v %v", ok, cheese)
	}

	val, ok = s1.Get("hat")
	hat := val.(int)
	if !ok || err != nil || hat != 3 {
		t.Fatalf("could not get right hat %v %v", ok, hat)
	}

	sh.Destroy(c, s1)
}

func TestSessionResetTTL(t *testing.T) {
	db, err := sql.Open("postgres", "postgres://test:test@localhost:/test?sslmode=disable")
	if err != nil {
		t.Skipf("Cannot connect to postgres. %v", err)
	}
	c := web.C{
		Env: map[interface{}]interface{}{"db": db},
	}
	sh, err := NewSessionHolder(db)
	if err != nil {
		t.Fatalf("failed to create session holder - %v", err)
	}

	// Create the session
	s := sh.Create(c)

	err = sh.Save(c, s)
	if err != nil {
		t.Fatalf("failed to save session - %v", err)
	}

	// Just be sure the call doesn't fail
	err = sh.ResetTTL(c, s)
	if err != nil {
		t.Fatalf("Failed to reset ttl - %v", err)
	}

	sh.Destroy(c, s)
}

func TestSessionCreate(t *testing.T) {
	db, err := sql.Open("postgres", "postgres://test:test@localhost:/test?sslmode=disable")
	if err != nil {
		t.Skipf("Cannot connect to postgres. %v", err)
	}
	c := web.C{
		Env: map[interface{}]interface{}{"db": db},
	}
	sh, err := NewSessionHolder(db)
	if err != nil {
		t.Fatalf("failed to create session holder - %v", err)
	}

	// Create the session
	s := sh.Create(c)

	s.Put("cheese", "cheddar")
	err = sh.Save(c, s)
	if err != nil {
		t.Fatalf("failed to save session - %v", err)
	}

	// Check we can write the session Id to the response
	w := httptest.NewRecorder()
	sh.AddToResponse(c, s, w)

	ch := w.HeaderMap.Get("Set-Cookie")
	if ch != fmt.Sprintf("sessionid=%s; Path=/; Max-Age=%d", s.Id(), sh.GetTimeout()) {
		t.Fatalf("cookie header is %v", ch)
	}

	// Can we extract the session from a request carrying a cookie?
	r, _ := http.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{
		Name:  "sessionid",
		Value: s.Id(),
	})

	s1, err := sh.Get(c, r)
	if err != nil {
		t.Fatalf("got error reading session from request, %v", err)
	}

	val, ok := s1.Get("cheese")
	cheese := val.(string)
	if !ok || cheese != "cheddar" {
		t.Fatalf("could not get right cheese %v %v", ok, cheese)
	}

	sh.Destroy(c, s1)

	s1, err = sh.Get(c, r)
	if err == nil {
		t.Fatalf("session should be destroyed")
	}
}
