package forgotPasswordHandler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/cache"
	"github.com/fossyy/filekeeper/view/client/email"
	"github.com/fossyy/filekeeper/view/client/forgotPassword"
	"github.com/google/uuid"
	"net/http"
	"sync"
	"time"

	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/types/models"
	"github.com/fossyy/filekeeper/utils"
	"gorm.io/gorm"
)

type ForgotPassword struct {
	User       *models.User
	Code       string
	mu         sync.Mutex
	CreateTime time.Time
}

var ListForgotPassword map[string]*ForgotPassword
var UserForgotPassword = make(map[string]string)

func init() {
	ListForgotPassword = make(map[string]*ForgotPassword)
	ticker := time.NewTicker(time.Minute)
	go func() {
		for {
			<-ticker.C
			currentTime := time.Now()
			cacheClean := 0
			cleanID := utils.GenerateRandomString(10)
			app.Server.Logger.Info(fmt.Sprintf("Cache cleanup [Forgot Password] [%s] initiated at %02d:%02d:%02d", cleanID, currentTime.Hour(), currentTime.Minute(), currentTime.Second()))

			for _, data := range ListForgotPassword {
				data.mu.Lock()
				if currentTime.Sub(data.CreateTime) > time.Minute*10 {
					delete(ListForgotPassword, data.User.Email)
					delete(UserForgotPassword, data.Code)
					cacheClean++
				}
				data.mu.Unlock()
			}

			app.Server.Logger.Info(fmt.Sprintf("Cache cleanup [Forgot Password] [%s] completed: %d entries removed. Finished at %s", cleanID, cacheClean, time.Since(currentTime)))
		}
	}()
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
}

func POST(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		app.Server.Logger.Error(err.Error())
		return
	}

	emailForm := r.Form.Get("email")

	user, err := cache.GetUser(emailForm)
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
	var code string

	var buffer bytes.Buffer
	data, ok := ListForgotPassword[user.Email]

	if !ok {
		code = utils.GenerateRandomString(64)
	} else {
		code = data.Code
	}

	err := emailView.ForgotPassword(user.Username, fmt.Sprintf("https://%s/forgot-password/verify/%s", utils.Getenv("DOMAIN"), code)).Render(context.Background(), &buffer)
	if err != nil {
		return err
	}

	userData := &ForgotPassword{
		User:       user,
		Code:       code,
		CreateTime: time.Now(),
	}

	UserForgotPassword[code] = user.Email
	ListForgotPassword[user.Email] = userData

	err = app.Server.Mail.Send(user.Email, "Password Change Request", buffer.String())
	if err != nil {
		return err
	}

	return nil
}
