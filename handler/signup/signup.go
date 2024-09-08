package signupHandler

import (
	"bytes"
	"context"
	"fmt"
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/utils"
	"github.com/fossyy/filekeeper/view/client/email"
	signupView "github.com/fossyy/filekeeper/view/client/signup"
	"net/http"
	"sync"
	"time"

	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/types/models"
	"github.com/google/uuid"
)

type UnverifiedUser struct {
	User       *models.User
	Code       string
	mu         sync.Mutex
	CreateTime time.Time
}

var VerifyUser map[string]*UnverifiedUser
var VerifyEmail map[string]string

func init() {
	VerifyUser = make(map[string]*UnverifiedUser)
	VerifyEmail = make(map[string]string)

	ticker := time.NewTicker(time.Minute)
	go func() {
		for {
			<-ticker.C
			currentTime := time.Now()
			cacheClean := 0
			cleanID := utils.GenerateRandomString(10)
			app.Server.Logger.Info(fmt.Sprintf("Cache cleanup [signup] [%s] initiated at %02d:%02d:%02d", cleanID, currentTime.Hour(), currentTime.Minute(), currentTime.Second()))

			for _, data := range VerifyUser {
				data.mu.Lock()
				if currentTime.Sub(data.CreateTime) > time.Minute*10 {
					delete(VerifyUser, data.Code)
					delete(VerifyEmail, data.User.Email)
					cacheClean++
				}
				data.mu.Unlock()
			}

			app.Server.Logger.Info(fmt.Sprintf("Cache cleanup [signup] [%s] completed: %d entries removed. Finished at %s", cleanID, cacheClean, time.Since(currentTime)))
		}
	}()
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

	code = VerifyEmail[user.Email]
	userData, ok := VerifyUser[code]

	if !ok {
		code = utils.GenerateRandomString(64)
	} else {
		code = userData.Code
	}

	err := emailView.RegistrationEmail(user.Username, fmt.Sprintf("https://%s/signup/verify/%s", utils.Getenv("DOMAIN"), code)).Render(context.Background(), &buffer)
	if err != nil {
		return err
	}

	unverifiedUser := UnverifiedUser{
		User:       user,
		Code:       code,
		CreateTime: time.Now(),
	}

	VerifyUser[code] = &unverifiedUser
	VerifyEmail[user.Email] = code

	err = app.Server.Mail.Send(user.Email, "Account Registration Verification", buffer.String())
	if err != nil {
		return err
	}
	return nil
}
