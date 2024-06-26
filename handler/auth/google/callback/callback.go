package googleOauthCallbackHandler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fossyy/filekeeper/cache"
	"github.com/fossyy/filekeeper/db"
	googleOauthSetupHandler "github.com/fossyy/filekeeper/handler/auth/google/setup"
	signinHandler "github.com/fossyy/filekeeper/handler/signin"
	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/session"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/utils"
	"net/http"
	"net/url"
	"sync"
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

type CsrfToken struct {
	Token      string
	CreateTime time.Time
	mu         sync.Mutex
}

var log *logger.AggregatedLogger
var CsrfTokens map[string]*CsrfToken

func init() {
	log = logger.Logger()
	CsrfTokens = make(map[string]*CsrfToken)

	ticker := time.NewTicker(time.Minute)
	go func() {
		for {
			<-ticker.C
			currentTime := time.Now()
			cacheClean := 0
			cleanID := utils.GenerateRandomString(10)
			log.Info(fmt.Sprintf("Cache cleanup [csrf_token] [%s] initiated at %02d:%02d:%02d", cleanID, currentTime.Hour(), currentTime.Minute(), currentTime.Second()))

			for _, data := range CsrfTokens {
				data.mu.Lock()
				if currentTime.Sub(data.CreateTime) > time.Minute*10 {
					delete(CsrfTokens, data.Token)
					cacheClean++
				}
				data.mu.Unlock()
			}

			log.Info(fmt.Sprintf("Cache cleanup [csrf_token] [%s] completed: %d entries removed. Finished at %s", cleanID, cacheClean, time.Since(currentTime)))
		}
	}()
}

func GET(w http.ResponseWriter, r *http.Request) {
	if _, ok := CsrfTokens[r.URL.Query().Get("state")]; !ok {
		//csrf token mismatch error
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	delete(CsrfTokens, r.URL.Query().Get("state"))

	if err := r.URL.Query().Get("error"); err != "" {
		http.Redirect(w, r, fmt.Sprintf("/signin?error=%s", err), http.StatusFound)
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
		log.Error("Error:", err)
		return
	}
	defer resp.Body.Close()

	var oauthData OauthToken
	if err := json.NewDecoder(resp.Body).Decode(&oauthData); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error("Error reading token response body:", err)
		return
	}

	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v1/userinfo?alt=json", nil)
	req.Header.Set("Authorization", "Bearer "+oauthData.AccessToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error("Error creating user info request:", err)
		return
	}

	userInfoResp, err := client.Do(req)
	defer userInfoResp.Body.Close()

	var oauthUser OauthUser
	if err := json.NewDecoder(userInfoResp.Body).Decode(&oauthUser); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error("Error reading user info response body:", err)
		return
	}

	if oauthUser.Email == "" {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error("Error reading user info response body: email not found")
		return
	}

	if !db.DB.IsEmailRegistered(oauthUser.Email) {
		code := utils.GenerateRandomString(64)
		googleOauthSetupHandler.SetupUser[code] = &googleOauthSetupHandler.UnregisteredUser{
			Code:       code,
			Email:      oauthUser.Email,
			CreateTime: time.Now(),
		}
		http.Redirect(w, r, fmt.Sprintf("/auth/google/setup/%s", code), http.StatusSeeOther)
		return
	}

	user, err := cache.GetUser(oauthUser.Email)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}

	storeSession := session.Create()
	storeSession.Values["user"] = types.User{
		UserID:        user.UserID,
		Email:         oauthUser.Email,
		Username:      user.Username,
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
	session.AddSessionInfo(oauthUser.Email, &sessionInfo)

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
