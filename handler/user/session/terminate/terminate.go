package userSessionTerminateHandler

import (
	"github.com/fossyy/filekeeper/session"
	userView "github.com/fossyy/filekeeper/view/user"
	"net/http"
)

func DELETE(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	_, mySession, _ := session.GetSession(r)
	otherSession, _ := session.Get(id)
	otherSession.Delete()
	session.RemoveSessionInfo(mySession.Email, otherSession.ID)

	component := userView.SessionTable(session.GetSessions(mySession.Email))
	err := component.Render(r.Context(), w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
