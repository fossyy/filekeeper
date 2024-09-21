package userSessionTerminateHandler

import (
	"errors"
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/session"
	"github.com/fossyy/filekeeper/types"
	"net/http"
)

func DELETE(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	mySession := r.Context().Value("user").(types.User)
	mySessionID := r.Context().Value("sessionID").(string)

	if id == mySessionID {
		w.Header().Set("HX-Redirect", "/logout")
		w.WriteHeader(http.StatusOK)
		return
	}

	otherSession := session.Get(id)
	err := session.RemoveSessionInfo(mySession.Email, otherSession.ID)
	if err != nil {
		if errors.Is(err, session.ErrorSessionNotFound) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	err = otherSession.Delete()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	return
}
