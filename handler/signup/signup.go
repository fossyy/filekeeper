package signupHandler

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/fossyy/filekeeper/db"
	"github.com/fossyy/filekeeper/email"
	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/types/models"
	"github.com/fossyy/filekeeper/utils"
	emailView "github.com/fossyy/filekeeper/view/email"
	signupView "github.com/fossyy/filekeeper/view/signup"
	"github.com/google/uuid"
)

type UnverifiedUser struct {
	User       *models.User
	Code       string
	mu         sync.Mutex
	CreateTime time.Time
}

var log *logger.AggregatedLogger
var mailServer *email.SmtpServer
var VerifyUser map[string]*UnverifiedUser
var VerifyEmail map[string]string

// TESTTING VAR
var database db.Database

func init() {
	log = logger.Logger()
	smtpPort, _ := strconv.Atoi(utils.Getenv("SMTP_PORT"))
	mailServer = email.NewSmtpServer(utils.Getenv("SMTP_HOST"), smtpPort, utils.Getenv("SMTP_USER"), utils.Getenv("SMTP_PASSWORD"))
	VerifyUser = make(map[string]*UnverifiedUser)
	VerifyEmail = make(map[string]string)
	database = db.NewMYSQLdb(utils.Getenv("DB_USERNAME"), utils.Getenv("DB_PASSWORD"), utils.Getenv("DB_HOST"), utils.Getenv("DB_PORT"), utils.Getenv("DB_NAME"))

	ticker := time.NewTicker(time.Minute)
	go func() {
		for {
			<-ticker.C
			currentTime := time.Now()
			cacheClean := 0
			log.Info(fmt.Sprintf("Cache cleanup initiated at %02d:%02d:%02d", currentTime.Hour(), currentTime.Minute(), currentTime.Second()))

			for _, data := range VerifyUser {
				data.mu.Lock()
				if currentTime.Sub(data.CreateTime) > time.Minute*1 {
					delete(VerifyUser, data.Code)
					delete(VerifyEmail, data.User.Email)
					cacheClean++
				}
				data.mu.Unlock()
			}

			log.Info(fmt.Sprintf("Cache cleanup completed: %d entries removed. Finished at %s", cacheClean, time.Since(currentTime)))
		}
	}()
}

func GET(w http.ResponseWriter, r *http.Request) {
	component := signupView.Main("Sign up Page", types.Message{
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}
	userEmail := r.Form.Get("email")
	username := r.Form.Get("username")
	password := r.Form.Get("password")
	isValid := utils.ValidatePassword(password)
	if !isValid {
		component := signupView.Main("Sign up Page", types.Message{
			Code:    0,
			Message: "Password is invalid",
		})
		err := component.Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error(err.Error())
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

	if registered := database.IsUserRegistered(userEmail, username); registered {
		component := signupView.Main("Sign up Page", types.Message{
			Code:    0,
			Message: "Email or Username has been registered",
		})
		err = component.Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error(err.Error())
			return
		}
		return
	}

	err = verifyEmail(&newUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}

	component := signupView.EmailSend("Sign up Page")
	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err.Error())
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

	err = mailServer.Send(user.Email, "Account Registration Verification", buffer.String())
	if err != nil {
		return err
	}
	return nil
}
