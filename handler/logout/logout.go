package logoutHandler

import (
	"errors"
	"github.com/fossyy/filekeeper/types"
	"net/http"

	"github.com/fossyy/filekeeper/session"
	"github.com/fossyy/filekeeper/utils"
)

func GET(w http.ResponseWriter, r *http.Request) {
	userSession := r.Context().Value("user").(types.User)
	cookie, err := r.Cookie("Session")
	if err != nil {
		return
	}
	storeSession := session.Get(cookie.Value)
	if err != nil {
		if errors.Is(err, &session.SessionNotFoundError{}) {
			storeSession.Destroy(w)
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = storeSession.Delete()
	if err != nil {
		panic(err)
		return
	}
	err = session.RemoveSessionInfo(userSession.Email, cookie.Value)
	if err != nil {
		panic(err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:   utils.Getenv("SESSION_NAME"),
		Value:  "",
		MaxAge: -1,
	})

	http.Redirect(w, r, "/signin", http.StatusSeeOther)
	return
}
