package admin

import (
	"github.com/fossyy/filekeeper/app"
	adminIndex "github.com/fossyy/filekeeper/view/admin/index"
	"net/http"
	"os"
	"path/filepath"
)

func SetupRoutes() *http.ServeMux {
	handler := http.NewServeMux()
	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//users, err := app.Admin.Database.GetAllUsers()
		//if err != nil {
		//	http.Error(w, "Unable to retrieve users", http.StatusInternalServerError)
		//	return
		//}
		//w.Header().Set("Content-Type", "application/json")
		//if err := json.NewEncoder(w).Encode(users); err != nil {
		//	http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		//	return
		//}
		adminIndex.Main().Render(r.Context(), w)
		return
	})
	handler.HandleFunc("/public/output.css", func(w http.ResponseWriter, r *http.Request) {
		openFile, err := os.OpenFile(filepath.Join("public", "output.css"), os.O_RDONLY, 0)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			app.Server.Logger.Error(err.Error())
			return
		}
		defer openFile.Close()
		stat, err := openFile.Stat()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			app.Server.Logger.Error(err.Error())
			return
		}
		http.ServeContent(w, r, openFile.Name(), stat.ModTime(), openFile)
	})
	return handler
}
