package signupVerifyHandler

import (
	"github.com/fossyy/filekeeper/app"
	signupView "github.com/fossyy/filekeeper/view/client/signup"
	"net/http"

	"github.com/fossyy/filekeeper/db"
	signupHandler "github.com/fossyy/filekeeper/handler/signup"
	"github.com/fossyy/filekeeper/types"
)

func GET(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	data, ok := signupHandler.VerifyUser[code]

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	err := db.DB.CreateUser(data.User)
	if err != nil {
		component := signupView.Main("Filekeeper - Sign up Page", types.Message{
			Code:    0,
			Message: "Email or Username has been registered",
		})
		err := component.Render(r.Context(), w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			app.Server.Logger.Error(err.Error())
			return
		}
		return
	}

	delete(signupHandler.VerifyUser, code)
	delete(signupHandler.VerifyEmail, data.User.Email)

	component := signupView.VerifySuccess("Filekeeper - Verify Page")

	err = component.Render(r.Context(), w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}
}
