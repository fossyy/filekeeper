package session

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/fossyy/filekeeper/utils"
)

type Session struct {
	ID     string
	Values map[string]interface{}
}

type SessionStore struct {
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

type SessionInfoList map[string][]*SessionInfo

var GlobalSessionStore = SessionStore{Sessions: make(map[string]*Session)}
var UserSessionInfoList = make(SessionInfoList)

type SessionNotFoundError struct{}

func (e *SessionNotFoundError) Error() string {
	return "session not found"
}

func (s *SessionStore) Get(id string) (*Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if session, ok := s.Sessions[id]; ok {
		return session, nil
	}
	return nil, &SessionNotFoundError{}
}

func (s *SessionStore) Create() *Session {
	id := utils.GenerateRandomString(128)
	session := &Session{
		ID:     id,
		Values: make(map[string]interface{}),
	}
	s.Sessions[id] = session
	return session
}

func (s *SessionStore) Delete(id string) {
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

func AddSessionInfo(email string, sessionInfo *SessionInfo) {
	UserSessionInfoList[email] = append(UserSessionInfoList[email], sessionInfo)
}

func RemoveSessionInfo(email string, id string) {
	sessionInfos := UserSessionInfoList[email]
	var updatedSessionInfos []*SessionInfo
	for _, sessionInfo := range sessionInfos {
		if sessionInfo.SessionID != id {
			updatedSessionInfos = append(updatedSessionInfos, sessionInfo)
		}
	}
	if len(updatedSessionInfos) > 0 {
		UserSessionInfoList[email] = updatedSessionInfos
		return
	}
	delete(UserSessionInfoList, email)
}

func RemoveAllSessions(email string) {
	sessionInfos := UserSessionInfoList[email]
	for _, sessionInfo := range sessionInfos {
		delete(GlobalSessionStore.Sessions, sessionInfo.SessionID)
	}
	delete(UserSessionInfoList, email)
}

func GetSessionInfo(email string, id string) *SessionInfo {
	for _, sessionInfo := range UserSessionInfoList[email] {
		if sessionInfo.SessionID == id {
			return sessionInfo
		}
	}
	return nil
}

func (sessionInfo *SessionInfo) UpdateAccessTime() {
	currentTime := time.Now()
	formattedTime := currentTime.Format("01-02-2006")
	sessionInfo.AccessAt = formattedTime
}
