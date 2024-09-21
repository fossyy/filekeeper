package forgotPasswordHandler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/view/client/email"
	"github.com/fossyy/filekeeper/view/client/forgotPassword"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"net/http"
	"time"

	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/types/models"
	"github.com/fossyy/filekeeper/utils"
	"gorm.io/gorm"
)

type ForgotPassword struct {
	User *models.User
	Code string
}

func GET(w http.ResponseWriter, r *http.Request) {
	component := forgotPasswordView.Main("Filekeeper - Forgot Password Page", types.Message{
		Code:    3,
		Message: "",
	})
	err := component.Render(r.Context(), w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}
	return
}

func POST(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		app.Server.Logger.Error(err.Error())
		return
	}

	emailForm := r.Form.Get("email")

	user, err := app.Server.Service.GetUser(r.Context(), emailForm)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			component := forgotPasswordView.Main("Filekeeper - Forgot Password Page", types.Message{
				Code:    0,
				Message: "Unexpected error has occurred",
			})
			err := component.Render(r.Context(), w)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				app.Server.Logger.Error(err.Error())
				return
			}
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	userData := &models.User{
		UserID:   uuid.UUID{},
		Username: user.Username,
		Email:    user.Email,
		Password: "",
	}

	err = verifyForgot(userData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	component := forgotPasswordView.EmailSend("Filekeeper - Forgot Password Page")
	err = component.Render(r.Context(), w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}
	return
}

func verifyForgot(user *models.User) error {
	var userData *ForgotPassword
	var code string
	var err error
	code, err = app.Server.Cache.GetCache(context.Background(), "ForgotPasswordCode:"+user.Email)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			code = utils.GenerateRandomString(64)
			userData = &ForgotPassword{
				User: user,
				Code: code,
			}

			newForgotUser, err := json.Marshal(userData)
			if err != nil {
				return err
			}
			err = app.Server.Cache.SetCache(context.Background(), "ForgotPasswordCode:"+user.Email, code, time.Minute*15)
			if err != nil {
				return err
			}
			err = app.Server.Cache.SetCache(context.Background(), "ForgotPassword:"+userData.Code, newForgotUser, time.Minute*15)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		storedCode, err := app.Server.Cache.GetCache(context.Background(), "ForgotPassword:"+code)
		err = json.Unmarshal([]byte(storedCode), &userData)
		if err != nil {
			return err
		}
	}

	var buffer bytes.Buffer
	err = emailView.ForgotPassword(user.Username, fmt.Sprintf("https://%s/forgot-password/verify/%s", utils.Getenv("DOMAIN"), code)).Render(context.Background(), &buffer)
	if err != nil {
		return err
	}

	err = app.Server.Mail.Send(user.Email, "Password Change Request", buffer.String())
	if err != nil {
		return err
	}

	return nil
}
