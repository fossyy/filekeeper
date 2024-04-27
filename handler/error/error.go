package errorHandler

import (
	"net/http"

	"github.com/fossyy/filekeeper/logger"
	errorView "github.com/fossyy/filekeeper/view/error"
)

var log *logger.AggregatedLogger

func init() {
	log = logger.Logger()
}

func ALL(w http.ResponseWriter, r *http.Request) {
	component := errorView.Main("Not Found")
	err := component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}
}
