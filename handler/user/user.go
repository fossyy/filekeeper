package userHandler

import (
	"github.com/a-h/templ"
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/view/client/user"
	"net/http"

	"github.com/fossyy/filekeeper/session"
)

var errorMessages = map[string]string{
	"password_not_match": "The passwords provided do not match. Please try again.",
}

func GET(w http.ResponseWriter, r *http.Request) {
	var component templ.Component
	userSession := r.Context().Value("user").(types.User)
	sessions, err := session.GetSessions(userSession.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := r.URL.Query().Get("error"); err != "" {
		message, ok := errorMessages[err]
		if !ok {
			message = "Unknown error occurred. Please contact support at bagas@fossy.my.id for assistance."
		}

		component = userView.Main("Filekeeper - User Page", userSession, sessions, types.Message{
			Code:    0,
			Message: message,
		})
	} else {
		component = userView.Main("Filekeeper - User Page", userSession, sessions, types.Message{
			Code:    1,
			Message: "",
		})
	}
	err = component.Render(r.Context(), w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}
}
