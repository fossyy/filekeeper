package miscHandler

import (
	"net/http"
)

func Robot(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/public/robots.txt", http.StatusSeeOther)
}

func Favicon(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/public/favicon.ico", http.StatusSeeOther)
}
