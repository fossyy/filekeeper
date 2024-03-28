package userHandler

import (
	"github.com/fossyy/filekeeper/middleware"
	userView "github.com/fossyy/filekeeper/view/user"
	"net/http"
)

func GET(w http.ResponseWriter, r *http.Request) {
	session, err := middleware.Store.Get(r, "session")
	userSession := middleware.GetUser(session)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	component := userView.Main("anjay mabar", userSession.Email, userSession.Username)
	component.Render(r.Context(), w)
}
