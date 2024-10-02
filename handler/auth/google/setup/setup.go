package googleOauthSetupHandler

import (
	"encoding/json"
	"errors"
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/session"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/types/models"
	"github.com/fossyy/filekeeper/utils"
	"github.com/fossyy/filekeeper/view/client/auth"
	signupView "github.com/fossyy/filekeeper/view/client/auth/signup"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"net/http"
	"time"
)

type UnregisteredUser struct {
	Code       string
	Email      string
	CreateTime time.Time
}

func GET(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	_, err := app.Server.Cache.GetCache(r.Context(), "GoogleSetup:"+code)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			http.Redirect(w, r, "/auth/signup", http.StatusSeeOther)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	component := authView.GoogleSetup("Filekeeper - Setup Page", types.Message{
		Code:    3,
		Message: "",
	})
	err = component.Render(r.Context(), w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}
}

func POST(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	cache, err := app.Server.Cache.GetCache(r.Context(), "GoogleSetup:"+code)

	if errors.Is(err, redis.Nil) {
		http.Error(w, "Unauthorized Action", http.StatusUnauthorized)
		return
	}

	var unregisteredUser UnregisteredUser
	err = json.Unmarshal([]byte(cache), &unregisteredUser)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	err = r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}
	username := r.Form.Get("username")
	password := r.Form.Get("password")

	isValid := utils.ValidatePassword(password)
	if !isValid {
		component := authView.GoogleSetup("Filekeeper - Setup Page", types.Message{
			Code:    0,
			Message: "Password is invalid",
		})
		err := component.Render(r.Context(), w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			app.Server.Logger.Error(err.Error())
			return
		}
		return
	}

	hashedPassword, err := utils.HashPassword(password)
	userID := uuid.New()
	newUser := models.User{
		UserID:   userID,
		Username: username,
		Email:    unregisteredUser.Email,
		Password: hashedPassword,
	}

	err = app.Server.Database.CreateUser(&newUser)
	if err != nil {
		app.Server.Logger.Error(err.Error())
		component := signupView.Main("Filekeeper - Sign up Page", types.Message{
			Code:    0,
			Message: "Email or Username has been registered",
		})
		err := component.Render(r.Context(), w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			app.Server.Logger.Error(err.Error())
			return
		}
		return
	}

	storeSession, err := session.Create(types.User{
		UserID:        userID,
		Email:         unregisteredUser.Email,
		Username:      username,
		Authenticated: true,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	userAgent := r.Header.Get("User-Agent")
	browserInfo, osInfo := utils.ParseUserAgent(userAgent)

	sessionInfo := session.SessionInfo{
		SessionID: storeSession.ID,
		Browser:   browserInfo["browser"],
		Version:   browserInfo["version"],
		OS:        osInfo["os"],
		OSVersion: osInfo["version"],
		IP:        utils.ClientIP(r),
		Location:  "Indonesia",
	}

	storeSession.Save(w)
	err = session.AddSessionInfo(unregisteredUser.Email, &sessionInfo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	http.Redirect(w, r, "/user", http.StatusSeeOther)
	return
}
