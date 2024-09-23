package userHandlerResetPassword

import (
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/session"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/utils"
	"net/http"
)

func POST(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}
	userSession := r.Context().Value("user").(types.User)
	currentPassword := r.Form.Get("currentPassword")
	password := r.Form.Get("password")
	user, err := app.Server.Service.GetUser(r.Context(), userSession.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	hashPassword, err := utils.HashPassword(password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	if !utils.CheckPasswordHash(currentPassword, user.Password) {
		http.Redirect(w, r, "/user?error=password_not_match", http.StatusSeeOther)
		return
	}

	err = app.Server.Database.UpdateUserPassword(user.Email, hashPassword)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	err = session.RemoveAllSessions(userSession.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	err = app.Server.Service.RemoveUserCache(r.Context(), userSession.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	http.Redirect(w, r, "/signin", http.StatusSeeOther)
	return
}
