package signupHandler

import (
	"github.com/fossyy/filekeeper/db"
	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/types/models"
	"github.com/fossyy/filekeeper/utils"
	signupView "github.com/fossyy/filekeeper/view/signup"
	"github.com/google/uuid"
	"net/http"
)

var log *logger.AggregatedLogger

func init() {
	log = logger.Logger()
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
	email := r.Form.Get("email")
	username := r.Form.Get("username")
	password := r.Form.Get("password")
	hashedPassword, err := utils.HashPassword(password)

	newUser := models.User{
		UserID:   uuid.New(),
		Username: username,
		Email:    email,
		Password: hashedPassword,
	}

	err = db.DB.Create(&newUser).Error

	if err != nil {
		component := signupView.Main("Sign up Page", types.Message{
			Code:    0,
			Message: "Username or Password has been registered",
		})
		err := component.Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error(err.Error())
			return
		}
		return
	}

	component := signupView.Main("Sign up Page", types.Message{
		Code:    1,
		Message: "User creation success",
	})
	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}
}
