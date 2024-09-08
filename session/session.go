package session

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/types"
	"github.com/redis/go-redis/v9"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fossyy/filekeeper/utils"
)

type Session struct {
	ID string
}

type SessionInfo struct {
	SessionID string
	Browser   string
	Version   string
	OS        string
	OSVersion string
	IP        string
	Location  string
}

type UserStatus string
type SessionNotFoundError struct{}

const (
	Authorized     UserStatus = "authorized"
	Unauthorized   UserStatus = "unauthorized"
	InvalidSession UserStatus = "invalid_session"
)

func (e *SessionNotFoundError) Error() string {
	return "session not found"
}

func Get(id string) *Session {
	return &Session{ID: id}
}

func Create(values types.User) (*Session, error) {
	id := utils.GenerateRandomString(128)

	sessionData, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}

	err = app.Server.Cache.SetCache(context.Background(), "Session:"+id, string(sessionData), time.Hour*24*7)
	if err != nil {
		return nil, err
	}

	return &Session{ID: id}, nil
}

func (s *Session) Change(user types.User) error {
	newSessionValue, err := json.Marshal(user)
	if err != nil {
		return err
	}
	err = app.Server.Cache.SetCache(context.Background(), "Session:"+s.ID, newSessionValue, time.Hour*24*7)
	if err != nil {
		return err
	}
	return nil
}

func (s *Session) Delete() error {
	err := app.Server.Cache.DeleteCache(context.Background(), "Session:"+s.ID)
	if err != nil {
		return err
	}
	return nil
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

func AddSessionInfo(email string, sessionInfo *SessionInfo) error {
	sessionInfoData, err := json.Marshal(sessionInfo)
	if err != nil {
		return err
	}

	err = app.Server.Cache.SetCache(context.Background(), "UserSessionInfo:"+email+":"+sessionInfo.SessionID, string(sessionInfoData), time.Hour*24*7)
	if err != nil {
		return err
	}

	return nil
}

func RemoveSessionInfo(email string, id string) error {
	key := "UserSessionInfo:" + email + ":" + id
	err := app.Server.Cache.DeleteCache(context.Background(), key)
	if err != nil {
		return err
	}
	return nil
}

func RemoveAllSessions(email string) error {
	pattern := "UserSessionInfo:" + email + ":*"
	keys, err := app.Server.Cache.GetKeys(context.Background(), pattern)
	if err != nil {
		return err
	}

	for _, key := range keys {
		parts := strings.Split(key, ":")
		sessionID := parts[2]

		err = app.Server.Cache.DeleteCache(context.Background(), "Session:"+sessionID)
		if err != nil {
			return err
		}
		err = app.Server.Cache.DeleteCache(context.Background(), key)
		if err != nil {
			return err
		}
	}

	return nil
}

func GetSessionInfo(email string, id string) (*SessionInfo, error) {
	key := "UserSessionInfo:" + email + ":" + id

	sessionInfoData, err := app.Server.Cache.GetCache(context.Background(), key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	var sessionInfo SessionInfo
	err = json.Unmarshal([]byte(sessionInfoData), &sessionInfo)
	if err != nil {
		return nil, err
	}

	return &sessionInfo, nil
}

func GetSession(r *http.Request) (UserStatus, types.User, string) {
	cookie, err := r.Cookie("Session")
	if err != nil {
		return Unauthorized, types.User{}, ""
	}

	sessionData, err := app.Server.Cache.GetCache(context.Background(), "Session:"+cookie.Value)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return InvalidSession, types.User{}, ""
		}
		return Unauthorized, types.User{}, ""
	}

	var storeSession types.User
	err = json.Unmarshal([]byte(sessionData), &storeSession)

	if err != nil {
		return Unauthorized, types.User{}, ""
	}

	if !storeSession.Authenticated && storeSession.Totp != "" {
		return Unauthorized, storeSession, cookie.Value
	}

	if !storeSession.Authenticated {
		return Unauthorized, types.User{}, ""
	}
	return Authorized, storeSession, cookie.Value
}

func GetSessions(email string) ([]*SessionInfo, error) {
	pattern := "UserSessionInfo:" + email + ":*"
	keys, err := app.Server.Cache.GetKeys(context.Background(), pattern)
	if err != nil {
		return nil, err
	}

	var sessions []*SessionInfo
	for _, key := range keys {
		sessionData, err := app.Server.Cache.GetCache(context.Background(), key)
		if err != nil {
			return nil, err
		}

		var sessionInfo SessionInfo
		err = json.Unmarshal([]byte(sessionData), &sessionInfo)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, &sessionInfo)
	}

	return sessions, nil
}
