package totpHandler

import (
	"fmt"
	"github.com/fossyy/filekeeper/session"
	"github.com/fossyy/filekeeper/types"
	totpView "github.com/fossyy/filekeeper/view/totp"
	"github.com/xlzd/gotp"
	"net/http"
	"time"
)

func GET(w http.ResponseWriter, r *http.Request) {
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
	_, user, key := session.GetSession(r)

	totp := gotp.NewDefaultTOTP(user.Totp)
	if totp.Verify(code, time.Now().Unix()) {
		storeSession, err := session.Get(key)
		if err != nil {
			return
		}
		fmt.Println(storeSession)
		storeSession.Values["user"] = types.User{
			UserID:        user.UserID,
			Email:         user.Email,
			Username:      user.Username,
			Totp:          "",
			Authenticated: true,
		}
		http.Redirect(w, r, "/user", http.StatusFound)
		return
	} else {
		fmt.Fprint(w, "wrong")
	}
}
