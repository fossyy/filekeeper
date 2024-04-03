package logoutHandler

import (
	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/session"
	"net/http"
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

	session.Store.Delete(cookie.Value)

	http.SetCookie(w, &http.Cookie{
		Name:   "Session",
		Value:  "",
		MaxAge: -1,
	})
	
	http.Redirect(w, r, "/signin", http.StatusSeeOther)
	return
}
