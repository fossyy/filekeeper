package googleOauthHandler

import (
	"encoding/json"
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/session"
	"github.com/fossyy/filekeeper/utils"
	"net/http"
	"net/url"
	"time"
)

type CsrfToken struct {
	Token      string
	CreateTime time.Time
}

func GET(w http.ResponseWriter, r *http.Request) {
	token, err := utils.GenerateCSRFToken()
	if err != nil {
		http.Error(w, "Unable to start Google sign-in", http.StatusInternalServerError)
		return
	}
	csrfToken := CsrfToken{
		Token:      token,
		CreateTime: time.Now(),
	}
	newCsrfToken, err := json.Marshal(csrfToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = app.Server.Cache.SetCache(r.Context(), "CsrfTokens:"+token, newCsrfToken, time.Minute*15)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: "oauth_state", Value: token, Path: "/auth/google/callback", MaxAge: 900,
		HttpOnly: true, Secure: session.IsHTTPS(r), SameSite: http.SameSiteLaxMode,
	})
	oauthURL, err := url.Parse("https://accounts.google.com/o/oauth2/auth")
	if err != nil {
		http.Error(w, "Unable to start Google sign-in", http.StatusInternalServerError)
		return
	}
	oauthURL.RawQuery = url.Values{
		"scope":         {"email profile"},
		"response_type": {"code"},
		"access_type":   {"offline"},
		"state":         {token},
		"redirect_uri":  {utils.Getenv("GOOGLE_CALLBACK")},
		"client_id":     {utils.Getenv("GOOGLE_CLIENT_ID")},
	}.Encode()
	http.Redirect(w, r, oauthURL.String(), http.StatusFound)
	return
}
