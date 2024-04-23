package signupHandler

import (
	"bytes"
	"context"
	"fmt"
	"github.com/fossyy/filekeeper/email"
	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/types/models"
	"github.com/fossyy/filekeeper/utils"
	emailView "github.com/fossyy/filekeeper/view/email"
	signupView "github.com/fossyy/filekeeper/view/signup"
	"github.com/google/uuid"
	"net/http"
)

var log *logger.AggregatedLogger
var mailServer *email.SmtpServer
var VerifyUser map[string]*models.User

func init() {
	log = logger.Logger()
	mailServer = email.NewSmtpServer("mail.fossy.my.id", 25, "test@fossy.my.id", "Test123456")
	VerifyUser = make(map[string]*models.User)
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
	hashedPassword, err := utils.HashPassword(password)

	newUser := models.User{
		UserID:   uuid.New(),
		Username: username,
		Email:    userEmail,
		Password: hashedPassword,
	}

	err = verifyEmail(&newUser)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}
}

func verifyEmail(user *models.User) error {
	var buffer bytes.Buffer
	id := utils.GenerateRandomString(64)
	err := emailView.RegistrationEmail(user.Username, fmt.Sprintf("https://filekeeper.fossy.my.id/verify/%s", id)).Render(context.Background(), &buffer)
	if err != nil {
		return err
	}

	VerifyUser[id] = user

	err = mailServer.Send(user.Email, "Account Registration Verification", buffer.String())
	if err != nil {
		return err
	}
	return nil
}
