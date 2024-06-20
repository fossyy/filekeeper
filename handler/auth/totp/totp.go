package totpHandler

import (
	"errors"
	"github.com/fossyy/filekeeper/session"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/utils"
	totpView "github.com/fossyy/filekeeper/view/totp"
	"github.com/xlzd/gotp"
	"net/http"
	"strings"
	"time"
)

func GET(w http.ResponseWriter, r *http.Request) {
	_, user, _ := session.GetSession(r)
	if user.Authenticated || user.Totp == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	component := totpView.Main("Filekeeper - 2FA Page", types.Message{
		Code:    1,
		Message: "",
	})
	err := component.Render(r.Context(), w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	return
}

func POST(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	code := r.Form.Get("code")
	_, user, key := session.GetSession(r)
	totp := gotp.NewDefaultTOTP(user.Totp)

	if totp.Verify(code, time.Now().Unix()) {
		storeSession, err := session.Get(key)
		if err != nil {
			return
		}
		storeSession.Values["user"] = types.User{
			UserID:        user.UserID,
			Email:         user.Email,
			Username:      user.Username,
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
		session.AddSessionInfo(user.Email, &sessionInfo)

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
	} else {
		component := totpView.Main("Filekeeper - 2FA Page", types.Message{
			Code:    0,
			Message: "Incorrect code. Please try again with the latest code from your authentication app.",
		})
		err := component.Render(r.Context(), w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

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
