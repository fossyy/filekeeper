package userHandler

import (
	"github.com/fossyy/filekeeper/types"
	"net/http"

	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/session"
	userView "github.com/fossyy/filekeeper/view/user"
)

var log *logger.AggregatedLogger

func init() {
	log = logger.Logger()
}

func GET(w http.ResponseWriter, r *http.Request) {
	userSession := r.Context().Value("user").(types.User)
	component := userView.Main("Filekeeper - User Page", userSession, session.GetSessions(userSession.Email))
	err := component.Render(r.Context(), w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}
}
