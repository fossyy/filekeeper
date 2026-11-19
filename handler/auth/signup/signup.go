package signupHandler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/utils"
	signupView "github.com/fossyy/filekeeper/view/client/auth/signup"
	"github.com/fossyy/filekeeper/view/client/email"
	"github.com/redis/go-redis/v9"
	"net/http"
	"net/mail"
	"time"

	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/types/models"
	"github.com/google/uuid"
)

type UnverifiedUser struct {
	User *models.User
	Code string
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
	return
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
	if !utils.ValidatePassword(password) {
		http.Error(w, "Password is invalid", http.StatusBadRequest)
		return
	}
	parsedEmail, emailErr := mail.ParseAddress(userEmail)
	if emailErr != nil || parsedEmail.Address != userEmail || len(username) == 0 || len(username) > 64 {
		http.Error(w, "Invalid account details", http.StatusBadRequest)
		return
	}
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		http.Error(w, "Unable to create account", http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

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
	if code == "" {
		return errors.New("invalid verification state")
	}

	err = emailView.RegistrationEmail(user.Username, fmt.Sprintf("https://%s/auth/signup/verify/%s", utils.Getenv("DOMAIN"), code)).Render(context.Background(), &buffer)
	if err != nil {
		return err
	}

	unverifiedUser := UnverifiedUser{
		User: user,
		Code: code,
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
