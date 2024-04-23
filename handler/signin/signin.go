package signinHandler

import (
	"errors"
	"github.com/fossyy/filekeeper/db/model/user"
	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/session"
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
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		log.Error(err.Error())
		return
	}
	email := r.Form.Get("email")
	password := r.Form.Get("password")
	userData, err := user.Get(email)
	if err != nil {
		component := signinView.Main("Sign in Page", types.Message{
			Code:    0,
			Message: "Incorrect Username or Password",
		})
		log.Error(err.Error())
		err = component.Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error(err.Error())
			return
		}
		return
	}

	if email == userData.Email && utils.CheckPasswordHash(password, userData.Password) {
		storeSession := session.Store.Create()
		storeSession.Values["user"] = types.User{
			UserID:        userData.UserID,
			Email:         email,
			Username:      userData.Username,
			Authenticated: true,
		}
		storeSession.Save(w)
		cookie, err := r.Cookie("redirect")
		if errors.Is(err, http.ErrNoCookie) {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:   "redirect",
			MaxAge: -1,
		})
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
