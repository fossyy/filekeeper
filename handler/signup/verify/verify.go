package signupVerifyHandler

import (
	"github.com/fossyy/filekeeper/db"
	signupHandler "github.com/fossyy/filekeeper/handler/signup"
	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/types"
	signupView "github.com/fossyy/filekeeper/view/signup"
	"net/http"
)

var log *logger.AggregatedLogger

func init() {
	log = logger.Logger()
}

func GET(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	user, ok := signupHandler.VerifyUser[id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	err := db.DB.Create(&user).Error

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
