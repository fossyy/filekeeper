package userHandler

import (
	"errors"
	"net/http"

	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/middleware"
	"github.com/fossyy/filekeeper/session"
	userView "github.com/fossyy/filekeeper/view/user"
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
	userSession := middleware.GetUser(storeSession)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}

	component := userView.Main("User Page", userSession.Email, userSession.Username, session.UserSessionInfoList[userSession.Email])
	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}
}
