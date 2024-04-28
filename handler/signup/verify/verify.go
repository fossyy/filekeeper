package signupVerifyHandler

import (
	"github.com/fossyy/filekeeper/utils"
	"net/http"

	"github.com/fossyy/filekeeper/db"
	signupHandler "github.com/fossyy/filekeeper/handler/signup"
	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/types"
	signupView "github.com/fossyy/filekeeper/view/signup"
)

var log *logger.AggregatedLogger

// TESTTING VAR
var database db.Database

func init() {
	log = logger.Logger()
	database = db.NewPostgresDB(utils.Getenv("DB_USERNAME"), utils.Getenv("DB_PASSWORD"), utils.Getenv("DB_HOST"), utils.Getenv("DB_PORT"), utils.Getenv("DB_NAME"))

}

func GET(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	data, ok := signupHandler.VerifyUser[code]

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	err := database.CreateUser(data.User)
	if err != nil {
		component := signupView.Main("Sign up Page", types.Message{
			Code:    0,
			Message: "Email or Username has been registered",
		})
		err := component.Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error(err.Error())
			return
		}
		return
	}

	delete(signupHandler.VerifyUser, code)
	delete(signupHandler.VerifyEmail, data.User.Email)

	component := signupView.VerifySuccess("Verify page")

	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}
}
