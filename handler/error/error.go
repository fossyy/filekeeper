package errorHandler

import (
	"fmt"
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/view/client/error"
	"net/http"
)

func NotFound(w http.ResponseWriter, r *http.Request) {
	component := errorView.NotFound("Not Found")
	err := component.Render(r.Context(), w)
	if err != nil {
		fmt.Fprint(w, err.Error())
		app.Server.Logger.Error(err.Error())
		return
	}
}

func InternalServerError(w http.ResponseWriter, r *http.Request) {
	component := errorView.InternalServerError("Internal Server Error")
	err := component.Render(r.Context(), w)
	if err != nil {
		fmt.Fprint(w, err.Error())
		app.Server.Logger.Error(err.Error())
		return
	}
}
