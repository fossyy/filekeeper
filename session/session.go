package session

import (
	"github.com/fossyy/filekeeper/utils"
	"net/http"
	"strconv"
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
var userSessions = make(map[string][]string)

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

func (s *Session) Save(w http.ResponseWriter) {
	maxAge, _ := strconv.Atoi(utils.Getenv("SESSION_MAX_AGE"))
	http.SetCookie(w, &http.Cookie{
		Name:   utils.Getenv("SESSION_NAME"),
		Value:  s.ID,
		MaxAge: maxAge,
	})
}

func (s *Session) Destroy(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   utils.Getenv("SESSION_NAME"),
		Value:  "",
		MaxAge: -1,
	})
}

func AppendSession(email string, session *Session) {
	userSessions[email] = append(userSessions[email], session.ID)
}

func RemoveSession(email string, id string) {
	sessions := userSessions[email]
	var updatedSessions []string
	for _, userSession := range sessions {
		if userSession != id {
			updatedSessions = append(updatedSessions, userSession)
		}
	}
	if len(updatedSessions) > 0 {
		userSessions[email] = updatedSessions
		return
	}
	delete(userSessions, email)
}

func RemoveAllSession(email string) {
	sessions := userSessions[email]
	for _, session := range sessions {
		delete(Store.Sessions, session)
	}
	delete(userSessions, email)
}

func Getses() map[string][]string {
	return userSessions
}
