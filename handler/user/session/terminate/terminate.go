package userSessionTerminateHandler

import (
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/session"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/view/client/user"
	"net/http"
)

func DELETE(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	mySession := r.Context().Value("user").(types.User)
	otherSession := session.Get(id)
	if _, err := session.GetSessionInfo(mySession.Email, otherSession.ID); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err := otherSession.Delete()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	err = session.RemoveSessionInfo(mySession.Email, otherSession.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	sessions, err := session.GetSessions(mySession.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}
	component := userView.SessionTable(sessions)

	err = component.Render(r.Context(), w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	return
}
