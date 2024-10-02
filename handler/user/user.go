package userHandler

import (
	"fmt"
	"github.com/a-h/templ"
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/session"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/utils"
	"github.com/fossyy/filekeeper/view/client/user"
	"net/http"
)

var errorMessages = map[string]string{
	"password_not_match": "The passwords provided do not match. Please try again.",
}

func GET(w http.ResponseWriter, r *http.Request) {
	userSession := r.Context().Value("user").(types.User)
	sessions, err := session.GetSessions(userSession.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	allowance, err := app.Server.Database.GetAllowance(userSession.UserID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	usage, err := app.Server.Service.CalculateUserStorageUsage(r.Context(), userSession.UserID.String())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	allowanceStats := &types.Allowance{
		AllowanceByte:        utils.ConvertFileSize(allowance.AllowanceByte),
		AllowanceUsedByte:    utils.ConvertFileSize(usage),
		AllowanceUsedPercent: fmt.Sprintf("%.2f", float64(usage)/float64(allowance.AllowanceByte)*100),
	}

	var component templ.Component
	if err := r.URL.Query().Get("error"); err != "" {
		message, ok := errorMessages[err]
		if !ok {
			message = "Unknown error occurred. Please contact support at bagas@fossy.my.id for assistance."
		}
		if r.Header.Get("hx-request") == "true" {
			component = userView.MainContent("Filekeeper - User Page", userSession, allowanceStats, sessions, types.Message{
				Code:    0,
				Message: message,
			})
		} else {
			component = userView.Main("Filekeeper - User Page", userSession, allowanceStats, sessions, types.Message{
				Code:    0,
				Message: message,
			})
		}
	} else {
		if r.Header.Get("hx-request") == "true" {
			component = userView.MainContent("Filekeeper - User Page", userSession, allowanceStats, sessions, types.Message{
				Code:    1,
				Message: "",
			})
		} else {
			component = userView.Main("Filekeeper - User Page", userSession, allowanceStats, sessions, types.Message{
				Code:    1,
				Message: "",
			})
		}
	}
	err = component.Render(r.Context(), w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}
}
