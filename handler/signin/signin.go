package signinHandler

import (
	"errors"
	"github.com/fossyy/filekeeper/cache"
	"net/http"
	"strings"

	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/session"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/utils"
	signinView "github.com/fossyy/filekeeper/view/signin"
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
	userData, err := cache.GetUser(email)
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
		storeSession := session.Create()
		storeSession.Values["user"] = types.User{
			UserID:        userData.UserID,
			Email:         email,
			Username:      userData.Username,
			Authenticated: true,
		}

		userAgent := r.Header.Get("User-Agent")
		browserInfo, osInfo := ParseUserAgent(userAgent)

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
		session.AddSessionInfo(email, &sessionInfo)

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

func ParseUserAgent(userAgent string) (map[string]string, map[string]string) {
	browserInfo := make(map[string]string)
	osInfo := make(map[string]string)
	if strings.Contains(userAgent, "Firefox") {
		browserInfo["browser"] = "Firefox"
		parts := strings.Split(userAgent, "Firefox/")
		if len(parts) > 1 {
			version := strings.Split(parts[1], " ")[0]
			browserInfo["version"] = version
		}
	} else if strings.Contains(userAgent, "Chrome") {
		browserInfo["browser"] = "Chrome"
		parts := strings.Split(userAgent, "Chrome/")
		if len(parts) > 1 {
			version := strings.Split(parts[1], " ")[0]
			browserInfo["version"] = version
		}
	} else {
		browserInfo["browser"] = "Unknown"
		browserInfo["version"] = "Unknown"
	}

	if strings.Contains(userAgent, "Windows") {
		osInfo["os"] = "Windows"
		parts := strings.Split(userAgent, "Windows ")
		if len(parts) > 1 {
			version := strings.Split(parts[1], ";")[0]
			osInfo["version"] = version
		}
	} else if strings.Contains(userAgent, "Macintosh") {
		osInfo["os"] = "Mac OS"
		parts := strings.Split(userAgent, "Mac OS X ")
		if len(parts) > 1 {
			version := strings.Split(parts[1], ";")[0]
			osInfo["version"] = version
		}
	} else if strings.Contains(userAgent, "Linux") {
		osInfo["os"] = "Linux"
		osInfo["version"] = "Unknown"
	} else if strings.Contains(userAgent, "Android") {
		osInfo["os"] = "Android"
		parts := strings.Split(userAgent, "Android ")
		if len(parts) > 1 {
			version := strings.Split(parts[1], ";")[0]
			osInfo["version"] = version
		}
	} else if strings.Contains(userAgent, "iPhone") || strings.Contains(userAgent, "iPad") || strings.Contains(userAgent, "iPod") {
		osInfo["os"] = "iOS"
		parts := strings.Split(userAgent, "OS ")
		if len(parts) > 1 {
			version := strings.Split(parts[1], " ")[0]
			osInfo["version"] = version
		}
	} else {
		osInfo["os"] = "Unknown"
		osInfo["version"] = "Unknown"
	}

	return browserInfo, osInfo
}
