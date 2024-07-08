package userHandlerResetPassword

import (
	"github.com/fossyy/filekeeper/cache"
	"github.com/fossyy/filekeeper/db"
	"github.com/fossyy/filekeeper/session"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/utils"
	"net/http"
)

func POST(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	userSession := r.Context().Value("user").(types.User)
	currentPassword := r.Form.Get("currentPassword")
	password := r.Form.Get("password")
	user, err := cache.GetUser(userSession.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	hashPassword, err := utils.HashPassword(password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !utils.CheckPasswordHash(currentPassword, user.Password) {
		http.Redirect(w, r, "/user?error=password_not_match", http.StatusSeeOther)
		return
	}

	err = db.DB.UpdateUserPassword(user.Email, hashPassword)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	session.RemoveAllSessions(userSession.Email)
	cache.DeleteUser(userSession.Email)

	http.Redirect(w, r, "/signin", http.StatusSeeOther)
	return
}
