package forgotPasswordVerifyHandler

import (
	"encoding/json"
	"errors"
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/handler/auth/forgotPassword"
	"github.com/fossyy/filekeeper/session"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/utils"
	"github.com/fossyy/filekeeper/view/client/auth/forgotPassword"
	signupView "github.com/fossyy/filekeeper/view/client/auth/signup"
	"github.com/redis/go-redis/v9"
	"net/http"
)

func GET(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")

	_, err := app.Server.Cache.GetCache(r.Context(), "ForgotPassword:"+code)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	component := forgotPasswordView.NewPasswordForm("Filekeeper - Forgot Password Page", types.Message{
		Code:    3,
		Message: "",
	})
	err = component.Render(r.Context(), w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}
	return
}

func POST(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")

	data, err := app.Server.Cache.GetCache(r.Context(), "ForgotPassword:"+code)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	var userData *forgotPasswordHandler.ForgotPassword

	err = json.Unmarshal([]byte(data), &userData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	err = r.ParseForm()
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

	err = app.Server.Database.UpdateUserPassword(userData.User.Email, hashedPassword)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	err = app.Server.Cache.DeleteCache(r.Context(), "ForgotPasswordCode:"+userData.User.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	err = app.Server.Cache.DeleteCache(r.Context(), "ForgotPassword:"+code)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	err = session.RemoveAllSessions(userData.User.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	err = app.Server.Cache.RemoveUserCache(r.Context(), userData.User.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	component := forgotPasswordView.ChangeSuccess("Filekeeper - Forgot Password Page")
	err = component.Render(r.Context(), w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}
	return
}
