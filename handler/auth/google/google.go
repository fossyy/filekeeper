package googleOauthHandler

import (
	"fmt"
	googleOauthCallbackHandler "github.com/fossyy/filekeeper/handler/auth/google/callback"
	"github.com/fossyy/filekeeper/utils"
	"net/http"
	"time"
)

func GET(w http.ResponseWriter, r *http.Request) {
	token, err := utils.GenerateCSRFToken()
	googleOauthCallbackHandler.CsrfTokens[token] = &googleOauthCallbackHandler.CsrfToken{Token: token, CreateTime: time.Now()}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("https://accounts.google.com/o/oauth2/auth?scope=email profile&response_type=code&access_type=offline&state=%s&redirect_uri=%s/auth/google/callback&client_id=%s", token, "http://localhost:8000", "324904877864-vrbqof7nea1l89316d26sk0s76105hc4.apps.googleusercontent.com"), http.StatusFound)
	return
}
