package indexHandler

import (
	"github.com/fossyy/filekeeper/view/index"
	"net/http"
)

func GET(w http.ResponseWriter, r *http.Request) {
	component := indexView.Main("main page")
	component.Render(r.Context(), w)
}
