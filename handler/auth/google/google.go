package googleOauthHandler

import (
	"encoding/json"
	"fmt"
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/utils"
	"net/http"
	"time"
)

type CsrfToken struct {
	Token      string
	CreateTime time.Time
}

func GET(w http.ResponseWriter, r *http.Request) {
	token, err := utils.GenerateCSRFToken()
	csrfToken := CsrfToken{
		Token:      token,
		CreateTime: time.Now(),
	}
	newCsrfToken, err := json.Marshal(csrfToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	app.Server.Cache.SetCache(r.Context(), "CsrfTokens:"+token, newCsrfToken, time.Minute*15)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}
	http.Redirect(w, r, fmt.Sprintf("https://accounts.google.com/o/oauth2/auth?scope=email profile&response_type=code&access_type=offline&state=%s&redirect_uri=%s&client_id=%s", token, utils.Getenv("GOOGLE_CALLBACK"), utils.Getenv("GOOGLE_CLIENT_ID")), http.StatusFound)
	return
}
