package totpHandler

import (
	"errors"
	"fmt"
	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/session"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/utils"
	totpView "github.com/fossyy/filekeeper/view/totp"
	"github.com/google/uuid"
	"github.com/xlzd/gotp"
	"net/http"
	"strings"
	"sync"
	"time"
)

type TotpInfo struct {
	ID         string
	UserID     uuid.UUID
	Secret     string
	Email      string
	Username   string
	CreateTime time.Time
	mu         sync.Mutex
}

var TotpInfoList = make(map[string]*TotpInfo)
var log *logger.AggregatedLogger

func init() {
	log = logger.Logger()

	ticker := time.NewTicker(time.Minute)
	go func() {
		for {
			<-ticker.C
			currentTime := time.Now()
			cacheClean := 0
			cleanID := utils.GenerateRandomString(10)
			log.Info(fmt.Sprintf("Cache cleanup [TOTP] [%s] initiated at %02d:%02d:%02d", cleanID, currentTime.Hour(), currentTime.Minute(), currentTime.Second()))

			for _, data := range TotpInfoList {
				data.mu.Lock()
				if currentTime.Sub(data.CreateTime) > time.Minute*10 {
					delete(TotpInfoList, data.ID)
					cacheClean++
				}
				data.mu.Unlock()
			}

			log.Info(fmt.Sprintf("Cache cleanup [TOTP] [%s] completed: %d entries removed. Finished at %s", cleanID, cacheClean, time.Since(currentTime)))
		}
	}()
}

func GET(w http.ResponseWriter, r *http.Request) {
	_, ok := TotpInfoList[r.PathValue("id")]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	component := totpView.Main("Filekeeper - 2FA Page")
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
	data, ok := TotpInfoList[r.PathValue("id")]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	fmt.Println(data)
	totp := gotp.NewDefaultTOTP(data.Secret)
	if totp.Verify(code, time.Now().Unix()) {
		storeSession := session.Create()
		storeSession.Values["user"] = types.User{
			UserID:        data.UserID,
			Email:         data.Email,
			Username:      data.Username,
			Totp:          "",
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
		session.AddSessionInfo(data.Email, &sessionInfo)

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
		fmt.Fprint(w, "wrong")
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
