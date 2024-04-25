package forgotPasswordHandler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/fossyy/filekeeper/db"
	"github.com/fossyy/filekeeper/email"
	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/types/models"
	"github.com/fossyy/filekeeper/utils"
	emailView "github.com/fossyy/filekeeper/view/email"
	forgotPasswordView "github.com/fossyy/filekeeper/view/forgotPassword"
	"gorm.io/gorm"
	"net/http"
	"sync"
	"time"
)

type ForgotPassword struct {
	User       *models.User
	Code       string
	mu         sync.Mutex
	CreateTime time.Time
}

var log *logger.AggregatedLogger
var mailServer *email.SmtpServer
var ListForgotPassword map[string]*ForgotPassword
var UserForgotPassword = make(map[string]string)

func init() {
	log = logger.Logger()
	ListForgotPassword = make(map[string]*ForgotPassword)
	mailServer = email.NewSmtpServer("mail.fossy.my.id", 25, "test@fossy.my.id", "Test123456")
	ticker := time.NewTicker(time.Minute)
	go func() {
		for {
			<-ticker.C
			currentTime := time.Now()
			cacheClean := 0
			log.Info(fmt.Sprintf("Cache cleanup initiated at %02d:%02d:%02d", currentTime.Hour(), currentTime.Minute(), currentTime.Second()))

			for _, data := range ListForgotPassword {
				data.mu.Lock()
				if currentTime.Sub(data.CreateTime) > time.Minute*1 {
					delete(ListForgotPassword, data.User.Email)
					delete(UserForgotPassword, data.Code)
					cacheClean++
				}
				data.mu.Unlock()
			}

			log.Info(fmt.Sprintf("Cache cleanup completed: %d entries removed. Finished at %s", cacheClean, time.Since(currentTime)))
		}
	}()
}

func GET(w http.ResponseWriter, r *http.Request) {
	component := forgotPasswordView.Main("Forgot Password Page", types.Message{
		Code:    3,
		Message: "",
	})
	err := component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}
}

func POST(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		log.Error(err.Error())
		return
	}

	emailForm := r.Form.Get("email")

	var user models.User
	err = db.DB.Table("users").Where("email = ?", emailForm).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		component := forgotPasswordView.Main(fmt.Sprintf("Account with this email address %s is not found", emailForm), types.Message{
			Code:    0,
			Message: "",
		})
		err := component.Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error(err.Error())
			return
		}
		return
	}

	err = verifyForgot(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}

	component := forgotPasswordView.EmailSend("Forgot Password Page")
	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err.Error())
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

	err = mailServer.Send(user.Email, "Password Change Request", buffer.String())
	if err != nil {
		return err
	}

	return nil
}
