package signupHandler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/utils"
	"github.com/fossyy/filekeeper/view/client/email"
	signupView "github.com/fossyy/filekeeper/view/client/signup"
	"github.com/redis/go-redis/v9"
	"net/http"
	"time"

	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/types/models"
	"github.com/google/uuid"
)

type UnverifiedUser struct {
	User       *models.User
	Code       string
	CreateTime time.Time
}

func GET(w http.ResponseWriter, r *http.Request) {
	component := signupView.Main("Filekeeper - Sign up Page", types.Message{
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
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}
	userEmail := r.Form.Get("email")
	username := r.Form.Get("username")
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

	newUser := models.User{
		UserID:   uuid.New(),
		Username: username,
		Email:    userEmail,
		Password: hashedPassword,
	}

	if registered := app.Server.Database.IsUserRegistered(userEmail, username); registered {
		component := signupView.Main("Filekeeper - Sign up Page", types.Message{
			Code:    0,
			Message: "Email or Username has been registered",
		})
		err = component.Render(r.Context(), w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			app.Server.Logger.Error(err.Error())
			return
		}
		return
	}

	err = verifyEmail(&newUser)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	component := signupView.EmailSend("Filekeeper - Sign up Page")
	err = component.Render(r.Context(), w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}
	return
}

func verifyEmail(user *models.User) error {
	var buffer bytes.Buffer
	var code string

	storedCode, err := app.Server.Cache.GetCache(context.Background(), "VerificationCode:"+user.Email)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			code = utils.GenerateRandomString(64)
		} else {
			return err
		}
	} else {
		code = storedCode
	}

	err = emailView.RegistrationEmail(user.Username, fmt.Sprintf("https://%s/signup/verify/%s", utils.Getenv("DOMAIN"), code)).Render(context.Background(), &buffer)
	if err != nil {
		return err
	}

	unverifiedUser := UnverifiedUser{
		User:       user,
		Code:       code,
		CreateTime: time.Now(),
	}
	newUnverifiedUser, err := json.Marshal(unverifiedUser)
	if err != nil {
		return err
	}

	err = app.Server.Cache.SetCache(context.Background(), "UnverifiedUser:"+code, newUnverifiedUser, 10*time.Minute)
	if err != nil {
		return err
	}

	err = app.Server.Cache.SetCache(context.Background(), "VerificationCode:"+user.Email, code, 10*time.Minute)
	if err != nil {
		return err
	}

	err = app.Server.Mail.Send(user.Email, "Account Registration Verification", buffer.String())
	if err != nil {
		return err
	}

	return nil
}
