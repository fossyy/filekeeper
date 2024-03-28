package errorHandler

import (
	errorView "github.com/fossyy/filekeeper/view/error"
	"net/http"
)

func ALL(w http.ResponseWriter, r *http.Request) {
	component := errorView.Main("Not Found")
	component.Render(r.Context(), w)
}
