package miscHandler

import (
	"net/http"
)

func Robot(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/public/robots.txt", http.StatusSeeOther)
}

func Favicon(w http.ResponseWriter, r *http.Request) {
	//currentDir, _ := os.Getwd()
	//fmt.Println(currentDir)
	//logo := "../../../favicon.ico"
	//basePath := filepath.Join(currentDir, "public")
	//logoPath := filepath.Join(basePath, logo)
	//fmt.Println(filepath.Dir(logoPath))
	//if filepath.Dir(logoPath) != basePath {
	//	log.Print("invalid logo path", logoPath)
	//	w.WriteHeader(500)
	//	return
	//}
	//http.ServeContent()
	http.Redirect(w, r, "/public/favicon.ico", http.StatusSeeOther)
}
