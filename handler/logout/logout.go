package logoutHandler

import (
	"errors"
	"net/http"

	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/session"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/utils"
)

var log *logger.AggregatedLogger

func init() {
	log = logger.Logger()
}

func GET(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("Session")
	if err != nil {
		return
	}

	storeSession, err := session.GlobalSessionStore.Get(cookie.Value)
	if err != nil {
		if errors.Is(err, &session.SessionNotFoundError{}) {
			storeSession.Destroy(w)
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.GlobalSessionStore.Delete(cookie.Value)
	session.RemoveSessionInfo(storeSession.Values["user"].(types.User).Email, cookie.Value)

	http.SetCookie(w, &http.Cookie{
		Name:   utils.Getenv("SESSION_NAME"),
		Value:  "",
		MaxAge: -1,
	})

	http.Redirect(w, r, "/signin", http.StatusSeeOther)
	return
}
