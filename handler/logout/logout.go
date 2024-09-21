package logoutHandler

import (
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/types"
	"net/http"

	"github.com/fossyy/filekeeper/session"
	"github.com/fossyy/filekeeper/utils"
)

func GET(w http.ResponseWriter, r *http.Request) {
	userSession := r.Context().Value("user").(types.User)
	cookie, err := r.Cookie("Session")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}
	storeSession := session.Get(cookie.Value)

	err = storeSession.Delete()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	err = session.RemoveSessionInfo(userSession.Email, cookie.Value)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
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
