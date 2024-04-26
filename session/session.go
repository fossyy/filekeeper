package session

import (
	"github.com/fossyy/filekeeper/utils"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Session struct {
	ID     string
	Values map[string]interface{}
}

type StoreSession struct {
	Sessions map[string]*Session
	mu       sync.Mutex
}

type SessionInfo struct {
	SessionID string
	Browser   string
	Version   string
	OS        string
	OSVersion string
	IP        string
	Location  string
	AccessAt  string
}

type ListSessionInfo map[string][]*SessionInfo

var Store = StoreSession{Sessions: make(map[string]*Session)}
var UserSessions = make(ListSessionInfo)

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

func AppendSession(email string, sessionInfo *SessionInfo) {
	UserSessions[email] = append(UserSessions[email], sessionInfo)
}

func RemoveSession(email string, id string) {
	sessions := UserSessions[email]
	var updatedSessions []*SessionInfo
	for _, userSession := range sessions {
		if userSession.SessionID != id {
			updatedSessions = append(updatedSessions, userSession)
		}
	}
	if len(updatedSessions) > 0 {
		UserSessions[email] = updatedSessions
		return
	}
	delete(UserSessions, email)
}

func RemoveAllSession(email string) {
	sessions := UserSessions[email]
	for _, session := range sessions {
		delete(Store.Sessions, session.SessionID)
	}
	delete(UserSessions, email)
}

func GetSessionInfo(email string, id string) *SessionInfo {
	for _, session := range UserSessions[email] {
		if session.SessionID == id {
			return session
		}
	}
	return nil
}

func (user *SessionInfo) UpdateAccessTime() {
	currentTime := time.Now()
	formattedTime := currentTime.Format("01-02-2006")
	user.AccessAt = formattedTime
}

func Getses() *ListSessionInfo {
	return &UserSessions
}
