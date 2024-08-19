package forgotPasswordVerifyHandler

import (
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/cache"
	forgotPasswordHandler "github.com/fossyy/filekeeper/handler/forgotPassword"
	"github.com/fossyy/filekeeper/session"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/utils"
	"github.com/fossyy/filekeeper/view/client/forgotPassword"
	signupView "github.com/fossyy/filekeeper/view/client/signup"
	"net/http"
)

func init() {

	//TESTING

}

func GET(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")

	email := forgotPasswordHandler.UserForgotPassword[code]
	_, ok := forgotPasswordHandler.ListForgotPassword[email]

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	component := forgotPasswordView.NewPasswordForm("Filekeeper - Forgot Password Page", types.Message{
		Code:    3,
		Message: "",
	})
	err := component.Render(r.Context(), w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}
}

func POST(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")

	email := forgotPasswordHandler.UserForgotPassword[code]
	data, ok := forgotPasswordHandler.ListForgotPassword[email]

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	password := r.Form.Get("password")
	isValid := utils.ValidatePassword(password)
	if !isValid {
		component := signupView.Main("Filekeeper - Sign up Page", types.Message{
			Code:    0,
			Message: "Password is invalid",
		})
		err := component.Render(r.Context(), w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			app.Server.Logger.Error(err.Error())
			return
		}
		return
	}
	hashedPassword, err := utils.HashPassword(password)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	err = app.Server.Database.UpdateUserPassword(data.User.Email, hashedPassword)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	delete(forgotPasswordHandler.ListForgotPassword, data.User.Email)
	delete(forgotPasswordHandler.UserForgotPassword, data.Code)

	session.RemoveAllSessions(data.User.Email)

	cache.DeleteUser(data.User.Email)

	component := forgotPasswordView.ChangeSuccess("Filekeeper - Forgot Password Page")
	err = component.Render(r.Context(), w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}
	return
}
