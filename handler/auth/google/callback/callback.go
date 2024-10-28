package googleOauthCallbackHandler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fossyy/filekeeper/app"
	googleOauthSetupHandler "github.com/fossyy/filekeeper/handler/auth/google/setup"
	"github.com/fossyy/filekeeper/session"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/utils"
	"github.com/redis/go-redis/v9"
	"net/http"
	"net/url"
	"time"
)

//type OauthToken struct {
//	AccessToken  string `json:"access_token"`
//	ExpiresIn    int    `json:"expires_in"`
//	RefreshToken string `json:"refresh_token"`
//	Scope        string `json:"scope"`
//	TokenType    string `json:"token_type"`
//	IdToken      string `json:"id_token"`
//}
//
//type OauthUser struct {
//	Id            string `json:"id"`
//	Email         string `json:"email"`
//	VerifiedEmail bool   `json:"verified_email"`
//	Name          string `json:"name"`
//	GivenName     string `json:"given_name"`
//	Picture       string `json:"picture"`
//	Locale        string `json:"locale"`
//}

type OauthToken struct {
	AccessToken string `json:"access_token"`
}

type OauthUser struct {
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
}

func GET(w http.ResponseWriter, r *http.Request) {
	_, err := app.Server.Cache.GetCache(r.Context(), "CsrfTokens:"+r.URL.Query().Get("state"))
	if err != nil {
		if errors.Is(err, redis.Nil) {
			http.Redirect(w, r, fmt.Sprintf("/auth/signin?error=%s", "csrf_token_error"), http.StatusFound)
			return
		}
		app.Server.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = app.Server.Cache.DeleteCache(r.Context(), "CsrfTokens:"+r.URL.Query().Get("state"))
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("/auth/signin?error=%s", "csrf_token_error"), http.StatusFound)
		return
	}

	if err := r.URL.Query().Get("error"); err != "" {
		http.Redirect(w, r, fmt.Sprintf("/auth/signin?error=%s", err), http.StatusFound)
		return
	}

	formData := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {r.URL.Query().Get("code")},
		"client_id":     {utils.Getenv("GOOGLE_CLIENT_ID")},
		"client_secret": {utils.Getenv("GOOGLE_CLIENT_SECRET")},
		"redirect_uri":  {utils.Getenv("GOOGLE_CALLBACK")},
	}

	resp, err := http.Post("https://oauth2.googleapis.com/token", "application/x-www-form-urlencoded", bytes.NewBufferString(formData.Encode()))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error("Error:", err)
		return
	}
	defer resp.Body.Close()

	var oauthData OauthToken
	if err := json.NewDecoder(resp.Body).Decode(&oauthData); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error("Error reading token response body:", err)
		return
	}

	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v1/userinfo?alt=json", nil)
	req.Header.Set("Authorization", "Bearer "+oauthData.AccessToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error("Error creating user info request:", err)
		return
	}

	userInfoResp, err := client.Do(req)
	defer userInfoResp.Body.Close()

	var oauthUser OauthUser
	if err := json.NewDecoder(userInfoResp.Body).Decode(&oauthUser); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error("Error reading user info response body:", err)
		return
	}

	if oauthUser.Email == "" {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error("Error reading user info response body: email not found")
		return
	}

	if !app.Server.Database.IsEmailRegistered(oauthUser.Email) {
		code := utils.GenerateRandomString(64)

		user := googleOauthSetupHandler.UnregisteredUser{
			Code:       code,
			Email:      oauthUser.Email,
			CreateTime: time.Now(),
		}
		newGoogleSetupJSON, _ := json.Marshal(user)
		err = app.Server.Cache.SetCache(r.Context(), "GoogleSetup:"+code, newGoogleSetupJSON, time.Minute*15)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			app.Server.Logger.Error(err.Error())
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/auth/google/setup/%s", code), http.StatusSeeOther)
		return
	}

	user, err := app.Server.Cache.GetUser(r.Context(), oauthUser.Email)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	storeSession, err := session.Create(types.User{
		UserID:        user.UserID,
		Email:         oauthUser.Email,
		Username:      user.Username,
		Authenticated: true,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
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
	err = session.AddSessionInfo(oauthUser.Email, &sessionInfo)
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
}
