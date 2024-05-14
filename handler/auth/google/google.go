package googleOauthHandler

import (
	"fmt"
	googleOauthCallbackHandler "github.com/fossyy/filekeeper/handler/auth/google/callback"
	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/utils"
	"net/http"
	"time"
)

var log *logger.AggregatedLogger

func init() {
	log = logger.Logger()
}

func GET(w http.ResponseWriter, r *http.Request) {
	token, err := utils.GenerateCSRFToken()
	googleOauthCallbackHandler.CsrfTokens[token] = &googleOauthCallbackHandler.CsrfToken{Token: token, CreateTime: time.Now()}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}
	http.Redirect(w, r, fmt.Sprintf("https://accounts.google.com/o/oauth2/auth?scope=email profile&response_type=code&access_type=offline&state=%s&redirect_uri=%s&client_id=%s", token, utils.Getenv("GOOGLE_CALLBACK"), utils.Getenv("GOOGLE_CLIENT_ID")), http.StatusFound)
	return
}
