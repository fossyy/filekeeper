package forgotPasswordVerifyHandler

import (
	"github.com/fossyy/filekeeper/db"
	"github.com/fossyy/filekeeper/db/model/user"
	forgotPasswordHandler "github.com/fossyy/filekeeper/handler/forgotPassword"
	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/session"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/utils"
	forgotPasswordView "github.com/fossyy/filekeeper/view/forgotPassword"
	signupView "github.com/fossyy/filekeeper/view/signup"

	"net/http"
)

var log *logger.AggregatedLogger

// TESTTING VAR
var database db.Database

func init() {
	log = logger.Logger()
	//TESTING
	database = db.NewPostgresDB(utils.Getenv("DB_USERNAME"), utils.Getenv("DB_PASSWORD"), utils.Getenv("DB_HOST"), utils.Getenv("DB_PORT"), utils.Getenv("DB_NAME"))

}

func GET(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")

	email := forgotPasswordHandler.UserForgotPassword[code]
	_, ok := forgotPasswordHandler.ListForgotPassword[email]

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	component := forgotPasswordView.NewPasswordForm("Forgot Password Page", types.Message{
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
	code := r.PathValue("code")

	email := forgotPasswordHandler.UserForgotPassword[code]
	data, ok := forgotPasswordHandler.ListForgotPassword[email]

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}

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

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}

	err = database.UpdateUserPassword(data.User.Email, hashedPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}

	delete(forgotPasswordHandler.ListForgotPassword, data.User.Email)
	delete(forgotPasswordHandler.UserForgotPassword, data.Code)

	session.RemoveAllSessions(data.User.Email)

	user.DeleteCache(data.User.Email)

	component := forgotPasswordView.ChangeSuccess("Forgot Password Page")
	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}
	return
}