package base

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zenazn/goji/web"
)

func TestSession(t *testing.T) {
	c := web.C{
		Env: make(map[string]interface{}, 0),
	}

	sh := NewBaseSessionHolder(30)

	s := sh.Create(c)

	if !s.IsDirty() {
		t.Fatalf("session should be created dirty")
	}

	if s != c.Env["session"] {
		t.Fatalf("session should be added to Env")
	}

	s.SetDirty(false)
	if s.IsDirty() {
		t.Fatalf("failed to clear dirty")
	}

	s.SetDirty(true)
	if !s.IsDirty() {
		t.Fatalf("Failed to set dirty")
	}

	s.SetDirty(false)
	s.Put("cheese", "cheddar")
	if !s.IsDirty() {
		t.Fatalf("Put should set dirty")
	}

	cheese, ok := s.Get("cheese")
	if !ok || cheese.(string) != "cheddar" {
		t.Fatalf("not the cheese we were hoping for. %v %v", ok, cheese)
	}

	_, ok = s.Get("hat")
	if ok {
		t.Fatalf("Get returned ok when value is not set")
	}

	s2 := sh.Create(c)
	if s.Id() == s2.Id() {
		t.Fatalf("session IDs not unique %s == %s ", s.Id(), s2.Id())
	}
}

func TestBaseSessionHolderAddToResponse(t *testing.T) {
	c := web.C{
		Env: make(map[string]interface{}, 0),
	}

	sh := NewBaseSessionHolder(30)

	s := sh.Create(c)

	w := httptest.NewRecorder()
	sh.AddToResponse(c, s, w)

	cookie := w.HeaderMap.Get("Set-Cookie")
	if cookie != fmt.Sprintf("sessionid=%s; Path=/; Max-Age=30", s.Id()) {
		t.Fatalf("Cookie not as expected - have %s", cookie)
	}
}

func TestBaseSessionHolderGetSessionId(t *testing.T) {
	c := web.C{
		Env: make(map[string]interface{}, 0),
	}

	sh := NewBaseSessionHolder(30)

	s := sh.Create(c)

	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Set("Cookie", fmt.Sprintf("sessionid=%s", s.Id()))
	sessionId := sh.GetSessionId(r)

	if sessionId != s.Id() {
		t.Fatalf("session ID extracted does not match.  %s != %s", sessionId, s.Id())
	}
}

func TestSessionMiddleware(t *testing.T) {
	c := web.C{
		Env: make(map[string]interface{}, 0),
	}

	sh := NewMemorySessionHolder(30)
	m := BuildSessionMiddleware(sh)

	// Create a session and ensure it is in the holder
	s := sh.Create(c)
	sh.Save(c, s)

	// Build a request referencing the session
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Set("Cookie", fmt.Sprintf("sessionid=%s", s.Id()))

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	w := httptest.NewRecorder()
	// Serve the request
	m(&c, h).ServeHTTP(w, r)

	if c.Env["session"].(*Session) != s {
		t.Fatalf("session not added to c.Env")
	}
	
	// store old sessionId
	oldSessionId := s.Id();
	// regenerate sessionId
	sh.RegenerateId(c, s)
	
	// test if sessionId changed
	if oldSessionId == s.Id() {
		t.Fatalf("session ID did not change after regeneration request")
	}

	// Build a request referencing the session with its new id
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Set("Cookie", fmt.Sprintf("sessionid=%s", s.Id()))

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	w := httptest.NewRecorder()
	// Serve the request
	m(&c, h).ServeHTTP(w, r)

	if c.Env["session"].(*Session) != s {
		t.Fatalf("session update not added to c.Env")
	}
	
}
