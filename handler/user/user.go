package userHandler

import (
	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/middleware"
	userView "github.com/fossyy/filekeeper/view/user"
	"net/http"
)

var log *logger.AggregatedLogger

func init() {
	log = logger.Logger()
}

func GET(w http.ResponseWriter, r *http.Request) {
	session, err := middleware.Store.Get(r, "session")
	userSession := middleware.GetUser(session)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}
	component := userView.Main("User Page", userSession.Email, userSession.Username)
	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}
}
