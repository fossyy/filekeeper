package totpHandler

import (
	"crypto/subtle"
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/session"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/utils"
	"github.com/fossyy/filekeeper/view/client/totp"
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
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}
	code := r.Form.Get("code")
	if len(code) != 6 || !isNumericCode(code) {
		http.Error(w, "Invalid code", http.StatusBadRequest)
		return
	}
	_, user, key := session.GetSession(r)
	if key == "" || user.Totp == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	decryptedSecret, err := app.Server.Encryption.Decrypt(user.Totp)
	if err != nil {
		http.Error(w, "Unable to verify code", http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}
	totp := gotp.NewDefaultTOTP(decryptedSecret)

	validCode := false
	for _, timestamp := range []int64{time.Now().Unix() - 30, time.Now().Unix(), time.Now().Unix() + 30} {
		if subtle.ConstantTimeCompare([]byte(totp.At(timestamp)), []byte(code)) == 1 {
			validCode = true
		}
	}
	if validCode {
		storeSession := session.Get(key)
		err := storeSession.Change(types.User{
			UserID:        user.UserID,
			Email:         user.Email,
			Username:      user.Username,
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

		storeSession.Save(r, w)
		err = session.AddSessionInfo(user.Email, &sessionInfo)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			app.Server.Logger.Error(err.Error())
			return
		}

		cookie, err := r.Cookie("redirect")
		if err != nil || !validRedirect(cookie.Value) {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "redirect",
			MaxAge:   -1,
			Path:     "/",
			HttpOnly: true,
			Secure:   session.IsHTTPS(r),
			SameSite: http.SameSiteLaxMode,
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

func validRedirect(value string) bool {
	return strings.HasPrefix(value, "/") && !strings.HasPrefix(value, "//") && !strings.Contains(value, `\`)
}

func isNumericCode(value string) bool {
	for _, digit := range value {
		if digit < '0' || digit > '9' {
			return false
		}
	}
	return true
}
