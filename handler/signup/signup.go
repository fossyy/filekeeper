package signupHandler

import (
	"github.com/fossyy/filekeeper/db"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/utils"
	signupView "github.com/fossyy/filekeeper/view/signup"
	"github.com/google/uuid"
	"net/http"
)

func GET(w http.ResponseWriter, r *http.Request) {
	component := signupView.Main("Sign up Page", types.Message{
		Code:    3,
		Message: "",
	})
	component.Render(r.Context(), w)
}

func POST(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}
	email := r.Form.Get("email")
	username := r.Form.Get("username")
	password := r.Form.Get("password")
	hashedPassword, err := utils.HashPassword(password)

	newUser := db.User{
		UserID:   uuid.New(),
		Username: username,
		Email:    email,
		Password: hashedPassword,
	}

	err = db.DB.Create(&newUser).Error

	if err != nil {
		component := signupView.Main("Sign up Page", types.Message{
			Code:    0,
			Message: "Username atau Email sudah terdaftar",
		})
		component.Render(r.Context(), w)
		return
	}

	component := signupView.Main("Sign up Page", types.Message{
		Code:    1,
		Message: "User creation success",
	})
	component.Render(r.Context(), w)
}
