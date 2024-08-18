package userHandler

import (
	"github.com/a-h/templ"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/view/client/user"
	"net/http"

	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/session"
)

var log *logger.AggregatedLogger

func init() {
	log = logger.Logger()
}

var errorMessages = map[string]string{
	"password_not_match": "The passwords provided do not match. Please try again.",
}

func GET(w http.ResponseWriter, r *http.Request) {
	var component templ.Component
	userSession := r.Context().Value("user").(types.User)

	if err := r.URL.Query().Get("error"); err != "" {
		message, ok := errorMessages[err]
		if !ok {
			message = "Unknown error occurred. Please contact support at bagas@fossy.my.id for assistance."
		}

		component = userView.Main("Filekeeper - User Page", userSession, session.GetSessions(userSession.Email), types.Message{
			Code:    0,
			Message: message,
		})
	} else {
		component = userView.Main("Filekeeper - User Page", userSession, session.GetSessions(userSession.Email), types.Message{
			Code:    1,
			Message: "",
		})
	}
	err := component.Render(r.Context(), w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}
}
