package indexHandler

import (
	"github.com/fossyy/filekeeper/session"
	"github.com/fossyy/filekeeper/view/client/index"
	"net/http"

	"github.com/fossyy/filekeeper/logger"
)

var log *logger.AggregatedLogger

func init() {
	log = logger.Logger()
}

func GET(w http.ResponseWriter, r *http.Request) {
	_, userSession, _ := session.GetSession(r)
	component := indexView.Main("Secure File Hosting - Filekeeper", userSession)
	err := component.Render(r.Context(), w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}
}
