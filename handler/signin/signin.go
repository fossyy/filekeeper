package signinHandler

import (
	"errors"
	"github.com/fossyy/filekeeper/db"
	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/middleware"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/utils"
	signinView "github.com/fossyy/filekeeper/view/signin"
	"net/http"
)

var log *logger.AggregatedLogger

func init() {
	log = logger.Logger()
}

func GET(w http.ResponseWriter, r *http.Request) {
	component := signinView.Main("Sign in Page", types.Message{
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
	session, _ := middleware.Store.Get(r, "session")
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		log.Error(err.Error())
		return
	}
	email := r.Form.Get("email")
	password := r.Form.Get("password")
	var userData db.User

	if err := db.DB.Table("users").Where("email = ?", email).First(&userData).Error; err != nil {
		component := signinView.Main("Sign in Page", types.Message{
			Code:    0,
			Message: "Database error : " + err.Error(),
		})
		err := component.Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error(err.Error())
			return
		}
	}
	if email == userData.Email && utils.CheckPasswordHash(password, userData.Password) {
		session.Values["user"] = types.User{
			UserID:        userData.UserID,
			Email:         email,
			Username:      userData.Username,
			Authenticated: true,
		}
		err = session.Save(r, w)
		cookie, err := r.Cookie("redirect")
		if errors.Is(err, http.ErrNoCookie) {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, cookie.Value, http.StatusSeeOther)
		return
	}
	component := signinView.Main("Sign in Page", types.Message{
		Code:    0,
		Message: "Incorrect Username or Password",
	})
	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}
}
