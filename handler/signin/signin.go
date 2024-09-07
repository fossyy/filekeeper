package signinHandler

import (
	"errors"
	"github.com/a-h/templ"
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/session"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/utils"
	"github.com/fossyy/filekeeper/view/client/signin"
	"net/http"
	"strings"
)

var errorMessages = make(map[string]string)

func init() {

}

func init() {
	errorMessages = map[string]string{
		"redirect_uri_mismatch":      "The redirect URI provided does not match the one registered with our service. Please contact the administrator for assistance.",
		"invalid_request":            "The request is missing required parameters or has invalid values. Please try again or contact support for assistance.",
		"access_denied":              "Access was denied. You may have declined the request for permission. Please try again if you wish to grant access.",
		"unauthorized_client":        "You are not authorized to make this request. Please contact support for further assistance.",
		"unsupported_response_type":  "The requested response type is not supported. Please try again with a supported response type.",
		"invalid_scope":              "The requested scope is invalid or unknown. Please try again or contact support for assistance.",
		"server_error":               "Our server encountered an unexpected error. Please try again later or contact support for assistance.",
		"temporarily_unavailable":    "Our server is currently undergoing maintenance. Please try again later.",
		"invalid_grant":              "The authorization code or refresh token provided is invalid. Please try again or contact support for assistance.",
		"invalid_client":             "The client identifier provided is invalid. Please check your credentials and try again.",
		"invalid_token":              "The access token provided is invalid. Please try again or contact support for assistance.",
		"insufficient_scope":         "You do not have sufficient privileges to perform this action. Please contact support for assistance.",
		"interaction_required":       "Interaction with the authorization server is required. Please try again.",
		"login_required":             "You need to log in again to proceed. Please try logging in again.",
		"account_selection_required": "Please select an account to proceed with the request.",
		"consent_required":           "Consent is required to proceed. Please provide consent to continue.",
		"csrf_token_error":           "The CSRF token is missing or invalid. Please refresh the page and try again.",
	}
}

func GET(w http.ResponseWriter, r *http.Request) {
	var component templ.Component
	if err := r.URL.Query().Get("error"); err != "" {
		message, ok := errorMessages[err]
		if !ok {
			message = "Unknown error occurred. Please contact support at bagas@fossy.my.id for assistance."
		}

		component = signinView.Main("Sign in Page", types.Message{
			Code:    0,
			Message: message,
		})
	} else {
		component = signinView.Main("Filekeeper - Sign in Page", types.Message{
			Code:    3,
			Message: "",
		})
	}

	err := component.Render(r.Context(), w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}
}

func POST(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		app.Server.Logger.Error(err.Error())
		return
	}
	email := r.Form.Get("email")
	password := r.Form.Get("password")
	userData, err := app.Server.Service.GetUser(r.Context(), email)
	if err != nil {
		component := signinView.Main("Filekeeper - Sign in Page", types.Message{
			Code:    0,
			Message: "Incorrect Username or Password",
		})
		app.Server.Logger.Error(err.Error())
		err = component.Render(r.Context(), w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			app.Server.Logger.Error(err.Error())
			return
		}
		return
	}

	if email == userData.Email && utils.CheckPasswordHash(password, userData.Password) {
		if userData.Totp != "" {

			storeSession, err := session.Create(types.User{
				UserID:        userData.UserID,
				Email:         email,
				Username:      userData.Username,
				Totp:          userData.Totp,
				Authenticated: false,
			})
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			storeSession.Save(w)
			http.Redirect(w, r, "/auth/totp", http.StatusSeeOther)
			return
		}

		storeSession, err := session.Create(types.User{
			UserID:        userData.UserID,
			Email:         email,
			Username:      userData.Username,
			Authenticated: true,
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
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
		session.AddSessionInfo(email, &sessionInfo)

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
	component := signinView.Main("Filekeeper - Sign in Page", types.Message{
		Code:    0,
		Message: "Incorrect Username or Password",
	})
	err = component.Render(r.Context(), w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
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
