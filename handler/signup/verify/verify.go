package signupVerifyHandler

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/fossyy/filekeeper/app"
	signupHandler "github.com/fossyy/filekeeper/handler/signup"
	signupView "github.com/fossyy/filekeeper/view/client/signup"
	"github.com/redis/go-redis/v9"
	"net/http"

	"github.com/fossyy/filekeeper/types"
)

func GET(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	userDataStr, err := app.Server.Cache.GetCache(context.Background(), "UnverifiedUser:"+code)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		app.Server.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var unverifiedUser signupHandler.UnverifiedUser
	err = json.Unmarshal([]byte(userDataStr), &unverifiedUser)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	err = app.Server.Database.CreateUser(unverifiedUser.User)
	if err != nil {
		component := signupView.Main("Filekeeper - Sign up Page", types.Message{
			Code:    0,
			Message: "Email or Username has already been registered",
		})
		err := component.Render(r.Context(), w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			app.Server.Logger.Error(err.Error())
			return
		}
		return
	}

	err = app.Server.Cache.DeleteCache(context.Background(), "UnverifiedUser:"+code)
	if err != nil {
		app.Server.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = app.Server.Cache.DeleteCache(context.Background(), "VerificationCode:"+unverifiedUser.User.Email)
	if err != nil {
		app.Server.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	component := signupView.VerifySuccess("Filekeeper - Verify Page")
	err = component.Render(r.Context(), w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}
	return
}
