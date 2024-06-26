package logoutHandler

import (
	"errors"
	"net/http"

	"github.com/fossyy/filekeeper/session"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/utils"
)

func GET(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("Session")
	if err != nil {
		return
	}

	storeSession, err := session.Get(cookie.Value)
	if err != nil {
		if errors.Is(err, &session.SessionNotFoundError{}) {
			storeSession.Destroy(w)
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	storeSession.Delete()
	session.RemoveSessionInfo(storeSession.Values["user"].(types.User).Email, cookie.Value)

	http.SetCookie(w, &http.Cookie{
		Name:   utils.Getenv("SESSION_NAME"),
		Value:  "",
		MaxAge: -1,
	})

	http.Redirect(w, r, "/signin", http.StatusSeeOther)
	return
}
