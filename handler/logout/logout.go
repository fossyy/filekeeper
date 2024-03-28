package logoutHandler

import (
	"github.com/fossyy/filekeeper/middleware"
	"github.com/fossyy/filekeeper/types"
	"net/http"
)

func GET(w http.ResponseWriter, r *http.Request) {
	session, err := middleware.Store.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Options.MaxAge = -1
	session.Values["user"] = types.User{}
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/signin", http.StatusSeeOther)
}
