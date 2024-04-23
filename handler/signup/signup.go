package signupHandler

import (
	"bytes"
	"context"
	"fmt"
	"github.com/fossyy/filekeeper/email"
	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/utils"
	emailView "github.com/fossyy/filekeeper/view/email"
	signupView "github.com/fossyy/filekeeper/view/signup"
	"net/http"
)

var log *logger.AggregatedLogger
var mailServer *email.SmtpServer

func init() {
	log = logger.Logger()
	mailServer = email.NewSmtpServer("mail.fossy.my.id", 25, "test@fossy.my.id", "Test123456")
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
	//password := r.Form.Get("password")
	//hashedPassword, err := utils.HashPassword(password)

	//newUser := models.User{
	//	UserID:   uuid.New(),
	//	Username: username,
	//	Email:    userEmail,
	//	Password: hashedPassword,
	//}

	err = verifyEmail(username, userEmail)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}

	//err = db.DB.Create(&newUser).Error
	//
	//if err != nil {
	//	component := signupView.Main("Sign up Page", types.Message{
	//		Code:    0,
	//		Message: "Username or Password has been registered",
	//	})
	//	err := component.Render(r.Context(), w)
	//	if err != nil {
	//		http.Error(w, err.Error(), http.StatusInternalServerError)
	//		log.Error(err.Error())
	//		return
	//	}
	//	return
	//}
	//
	//component := signupView.Main("Sign up Page", types.Message{
	//	Code:    1,
	//	Message: "User creation success",
	//})
	//err = component.Render(r.Context(), w)
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	log.Error(err.Error())
	//	return
	//}
}

func verifyEmail(username string, email string) error {
	var buffer bytes.Buffer
	err := emailView.RegistrationEmail(username, fmt.Sprintf("https://filekeeper.fossy.my.id/verify/%s", utils.GenerateRandomString(64))).Render(context.Background(), &buffer)
	if err != nil {
		return err
	}
	err = mailServer.Send(email, "Account Registration Verification", buffer.String())
	if err != nil {
		return err
	}
	return nil
}
