package totpHandler

import (
	"errors"
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/session"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/utils"
	"github.com/fossyy/filekeeper/view/client/totp"
	"github.com/xlzd/gotp"
	"net/http"
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
	_, user, key := session.GetSession(r)
	totp := gotp.NewDefaultTOTP(user.Totp)

	if totp.Verify(code, time.Now().Unix()) {
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

		storeSession.Save(w)
		err = session.AddSessionInfo(user.Email, &sessionInfo)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			app.Server.Logger.Error(err.Error())
			return
		}

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
