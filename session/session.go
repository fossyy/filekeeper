package session

import (
	"errors"
	"github.com/fossyy/filekeeper/types"
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

type UserStatus string

const (
	Authorized     UserStatus = "authorized"
	Unauthorized   UserStatus = "unauthorized"
	InvalidSession UserStatus = "invalid_session"
)

var GlobalSessionStore = SessionStore{Sessions: make(map[string]*Session)}
var UserSessionInfoList = make(map[string]map[string]*SessionInfo)

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
		Path:   "/",
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
	if _, ok := UserSessionInfoList[email]; !ok {
		UserSessionInfoList[email] = make(map[string]*SessionInfo)
	}

	UserSessionInfoList[email][sessionInfo.SessionID] = sessionInfo
}

func RemoveSessionInfo(email string, id string) {
	if userSessions, ok := UserSessionInfoList[email]; ok {
		if _, ok := userSessions[id]; ok {
			delete(userSessions, id)
			if len(userSessions) == 0 {
				delete(UserSessionInfoList, email)
			}
		}
	}
}

func RemoveAllSessions(email string) {
	sessionInfos := UserSessionInfoList[email]
	for _, sessionInfo := range sessionInfos {
		delete(GlobalSessionStore.Sessions, sessionInfo.SessionID)
	}
	delete(UserSessionInfoList, email)
}

func GetSessionInfo(email string, id string) *SessionInfo {
	if userSession, ok := UserSessionInfoList[email]; ok {
		if sessionInfo, ok := userSession[id]; ok {
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

func GetSession(r *http.Request) (UserStatus, types.User, string) {
	cookie, err := r.Cookie("Session")
	if err != nil {
		return Unauthorized, types.User{}, ""
	}

	storeSession, err := GlobalSessionStore.Get(cookie.Value)
	if err != nil {
		if errors.Is(err, &SessionNotFoundError{}) {
			return InvalidSession, types.User{}, ""
		}
		return Unauthorized, types.User{}, ""
	}

	val := storeSession.Values["user"]
	var userSession = types.User{}
	userSession, ok := val.(types.User)
	if !ok {
		return Unauthorized, types.User{}, ""
	}

	return Authorized, userSession, cookie.Value
}

func GetSessions(email string) []*SessionInfo {
	if sessions, ok := UserSessionInfoList[email]; ok {
		result := make([]*SessionInfo, 0, len(sessions))
		for _, sessionInfo := range sessions {
			result = append(result, sessionInfo)
		}
		return result
	}
	return nil
}
