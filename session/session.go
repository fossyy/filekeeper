package session

import (
	"github.com/fossyy/filekeeper/utils"
	"sync"
)

type Session struct {
	ID     string
	Values map[string]interface{}
}

type StoreSession struct {
	Sessions map[string]*Session
	mu       sync.Mutex
}

var Store = StoreSession{Sessions: make(map[string]*Session)}

type SessionNotFound struct{}

func (e *SessionNotFound) Error() string {
	return "session not found"
}

func (s *StoreSession) Get(id string) (*Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if session, ok := s.Sessions[id]; ok {
		return session, nil
	}
	return nil, &SessionNotFound{}
}

func (s *StoreSession) Create() *Session {
	id := utils.GenerateRandomString(128)
	session := &Session{
		ID:     id,
		Values: make(map[string]interface{}),
	}
	s.Sessions[id] = session
	return session
}

func (s *StoreSession) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.Sessions, id)
}
