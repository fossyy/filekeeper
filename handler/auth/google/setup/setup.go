package googleOauthSetupHandler

import (
	"fmt"
	"github.com/fossyy/filekeeper/db"
	signinHandler "github.com/fossyy/filekeeper/handler/signin"
	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/session"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/types/models"
	"github.com/fossyy/filekeeper/utils"
	authView "github.com/fossyy/filekeeper/view/auth"
	signupView "github.com/fossyy/filekeeper/view/signup"
	"github.com/google/uuid"
	"net/http"
	"sync"
	"time"
)

type UnregisteredUser struct {
	Code       string
	Email      string
	CreateTime time.Time
	mu         sync.Mutex
}

var log *logger.AggregatedLogger
var SetupUser map[string]*UnregisteredUser

func init() {
	log = logger.Logger()
	SetupUser = make(map[string]*UnregisteredUser)

	ticker := time.NewTicker(time.Minute)
	go func() {
		for {
			<-ticker.C
			currentTime := time.Now()
			cacheClean := 0
			cleanID := utils.GenerateRandomString(10)
			log.Info(fmt.Sprintf("Cache cleanup [GoogleSetup] [%s] initiated at %02d:%02d:%02d", cleanID, currentTime.Hour(), currentTime.Minute(), currentTime.Second()))

			for _, data := range SetupUser {
				data.mu.Lock()
				if currentTime.Sub(data.CreateTime) > time.Minute*10 {
					delete(SetupUser, data.Code)
					cacheClean++
				}

				data.mu.Unlock()
			}

			log.Info(fmt.Sprintf("Cache cleanup [GoogleSetup] [%s] completed: %d entries removed. Finished at %s", cleanID, cacheClean, time.Since(currentTime)))
		}
	}()
}

func GET(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if _, ok := SetupUser[code]; !ok {
		http.Redirect(w, r, "/signup", http.StatusSeeOther)
		return
	}
	component := authView.GoogleSetup("Filekeeper - Setup Page", types.Message{
		Code:    3,
		Message: "",
	})
	err := component.Render(r.Context(), w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}
}

func POST(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	unregisteredUser, ok := SetupUser[code]
	if !ok {
		http.Error(w, "Unauthorized Action", http.StatusUnauthorized)
		return
	}
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error(err.Error())
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
			log.Error(err.Error())
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

	err = db.DB.CreateUser(&newUser)
	if err != nil {
		component := signupView.Main("Filekeeper - Sign up Page", types.Message{
			Code:    0,
			Message: "Email or Username has been registered",
		})
		err := component.Render(r.Context(), w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Error(err.Error())
			return
		}
		return
	}

	delete(SetupUser, code)

	storeSession := session.Create()
	storeSession.Values["user"] = types.User{
		UserID:        userID,
		Email:         unregisteredUser.Email,
		Username:      username,
		Authenticated: true,
	}

	userAgent := r.Header.Get("User-Agent")
	browserInfo, osInfo := signinHandler.ParseUserAgent(userAgent)

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
	session.AddSessionInfo(unregisteredUser.Email, &sessionInfo)

	http.Redirect(w, r, "/user", http.StatusSeeOther)
	return
}
