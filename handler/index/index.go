package indexHandler

import (
	"github.com/fossyy/filekeeper/session"
	"net/http"

	"github.com/fossyy/filekeeper/logger"
	indexView "github.com/fossyy/filekeeper/view/index"
)

var log *logger.AggregatedLogger

func init() {
	log = logger.Logger()
}

func GET(w http.ResponseWriter, r *http.Request) {
	_, userSession, _ := session.GetSession(r)
	component := indexView.Main("main page", userSession)
	err := component.Render(r.Context(), w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}
}
