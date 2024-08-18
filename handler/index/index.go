package indexHandler

import (
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/session"
	"github.com/fossyy/filekeeper/view/client/index"
	"net/http"
)

func GET(w http.ResponseWriter, r *http.Request) {
	_, userSession, _ := session.GetSession(r)
	component := indexView.Main("Secure File Hosting - Filekeeper", userSession)
	err := component.Render(r.Context(), w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}
}
