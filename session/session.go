package session

import (
	"fmt"
	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/types"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/fossyy/filekeeper/utils"
)

type Session struct {
	ID         string
	Values     map[string]interface{}
	CreateTime time.Time
	mu         sync.Mutex
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
type SessionNotFoundError struct{}

const (
	Authorized     UserStatus = "authorized"
	Unauthorized   UserStatus = "unauthorized"
	InvalidSession UserStatus = "invalid_session"
)

var GlobalSessionStore = make(map[string]*Session)
var UserSessionInfoList = make(map[string]map[string]*SessionInfo)
var log *logger.AggregatedLogger

func init() {
	log = logger.Logger()

	ticker := time.NewTicker(time.Minute)
	go func() {
		for {
			<-ticker.C
			currentTime := time.Now()
			cacheClean := 0
			cleanID := utils.GenerateRandomString(10)
			log.Info(fmt.Sprintf("Cache cleanup [Session] [%s] initiated at %02d:%02d:%02d", cleanID, currentTime.Hour(), currentTime.Minute(), currentTime.Second()))

			for _, data := range GlobalSessionStore {
				data.mu.Lock()
				if currentTime.Sub(data.CreateTime) > time.Hour*24*7 {
					RemoveSessionInfo(data.Values["user"].(types.User).Email, data.ID)
					delete(GlobalSessionStore, data.ID)
					cacheClean++
				}
				data.mu.Unlock()
			}

			log.Info(fmt.Sprintf("Cache cleanup [Session] [%s] completed: %d entries removed. Finished at %s", cleanID, cacheClean, time.Since(currentTime)))
		}
	}()
}

func (e *SessionNotFoundError) Error() string {
	return "session not found"
}

func Get(id string) (*Session, error) {
	if session, ok := GlobalSessionStore[id]; ok {
		return session, nil
	}
	return nil, &SessionNotFoundError{}
}

func Create() *Session {
	id := utils.GenerateRandomString(128)
	session := &Session{
		ID:     id,
		Values: make(map[string]interface{}),
	}
	GlobalSessionStore[id] = session
	return session
}

func (s *Session) Delete() {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(GlobalSessionStore, s.ID)
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
		delete(GlobalSessionStore, sessionInfo.SessionID)
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

	storeSession, ok := GlobalSessionStore[cookie.Value]
	if !ok {
		return InvalidSession, types.User{}, ""
	}

	val := storeSession.Values["user"]
	var userSession = types.User{}
	userSession, ok = val.(types.User)
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
