package main

import (
	"fmt"
	downloadHandler "github.com/fossyy/filekeeper/handler/download"
	downloadFileHandler "github.com/fossyy/filekeeper/handler/download/file"
	indexHandler "github.com/fossyy/filekeeper/handler/index"
	logoutHandler "github.com/fossyy/filekeeper/handler/logout"
	miscHandler "github.com/fossyy/filekeeper/handler/misc"
	signinHandler "github.com/fossyy/filekeeper/handler/signin"
	signupHandler "github.com/fossyy/filekeeper/handler/signup"
	uploadHandler "github.com/fossyy/filekeeper/handler/upload"
	"github.com/fossyy/filekeeper/handler/upload/initialisation"
	userHandler "github.com/fossyy/filekeeper/handler/user"
	"github.com/fossyy/filekeeper/middleware"
	"github.com/fossyy/filekeeper/utils"
	"net/http"
)

func main() {
	handler := http.NewServeMux()
	serverAddr := fmt.Sprintf("%s:%s", utils.Getenv("SERVER_HOST"), utils.Getenv("SERVER_PORT"))
	server := http.Server{
		Addr:    serverAddr,
		Handler: middleware.Handler(handler),
	}

	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			indexHandler.GET(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	handler.HandleFunc("/signin", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			middleware.Guest(signinHandler.GET, w, r)
		case http.MethodPost:
			middleware.Guest(signinHandler.POST, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	handler.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			middleware.Guest(signupHandler.GET, w, r)
		case http.MethodPost:
			middleware.Guest(signupHandler.POST, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	handler.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			middleware.Auth(userHandler.GET, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	handler.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			middleware.Auth(uploadHandler.GET, w, r)
		case http.MethodPost:
			middleware.Auth(uploadHandler.POST, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	handler.HandleFunc("/upload/init", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			middleware.Auth(initialisation.POST, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	handler.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			middleware.Auth(downloadHandler.GET, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	handler.HandleFunc("/download/{id}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			downloadFileHandler.GET(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	handler.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		middleware.Auth(logoutHandler.GET, w, r)
	})

	handler.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		miscHandler.Robot(w, r)
	})

	handler.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {

		http.Redirect(w, r, "/public/favicon.ico", http.StatusSeeOther)
	})

	fileServer := http.FileServer(http.Dir("./public"))
	handler.Handle("/public/", http.StripPrefix("/public", fileServer))

	fmt.Printf("Listening on http://%s\n", serverAddr)
	err := server.ListenAndServe()
	if err != nil {
		return
	}
}
