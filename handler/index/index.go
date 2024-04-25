package indexHandler

import (
	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/view/index"
	"net/http"
)

var log *logger.AggregatedLogger

func init() {
	log = logger.Logger()
}

func GET(w http.ResponseWriter, r *http.Request) {
	component := indexView.Main("main page")
	err := component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}
}
