package userSessionTerminateHandler

import (
	"github.com/fossyy/filekeeper/session"
	"github.com/fossyy/filekeeper/view/client/user"
	"net/http"
)

func DELETE(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	_, mySession, _ := session.GetSession(r)
	otherSession, _ := session.Get(id)
	if _, err := session.GetSessionInfo(mySession.Email, otherSession.ID); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	otherSession.Delete()
	session.RemoveSessionInfo(mySession.Email, otherSession.ID)
	sessions, err := session.GetSessions(mySession.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	component := userView.SessionTable(sessions)

	err = component.Render(r.Context(), w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
