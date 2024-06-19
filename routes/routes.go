package routes

import (
	googleOauthHandler "github.com/fossyy/filekeeper/handler/auth/google"
	googleOauthCallbackHandler "github.com/fossyy/filekeeper/handler/auth/google/callback"
	googleOauthSetupHandler "github.com/fossyy/filekeeper/handler/auth/google/setup"
	totpHandler "github.com/fossyy/filekeeper/handler/auth/totp"
	downloadHandler "github.com/fossyy/filekeeper/handler/download"
	downloadFileHandler "github.com/fossyy/filekeeper/handler/download/file"
	forgotPasswordHandler "github.com/fossyy/filekeeper/handler/forgotPassword"
	forgotPasswordVerifyHandler "github.com/fossyy/filekeeper/handler/forgotPassword/verify"
	indexHandler "github.com/fossyy/filekeeper/handler/index"
	logoutHandler "github.com/fossyy/filekeeper/handler/logout"
	signinHandler "github.com/fossyy/filekeeper/handler/signin"
	signupHandler "github.com/fossyy/filekeeper/handler/signup"
	signupVerifyHandler "github.com/fossyy/filekeeper/handler/signup/verify"
	uploadHandler "github.com/fossyy/filekeeper/handler/upload"
	"github.com/fossyy/filekeeper/handler/upload/initialisation"
	userHandler "github.com/fossyy/filekeeper/handler/user"
	userHandlerTotpSetup "github.com/fossyy/filekeeper/handler/user/totp"
	"github.com/fossyy/filekeeper/middleware"
	"net/http"
)

func SetupRoutes() *http.ServeMux {
	handler := http.NewServeMux()

	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case "/":
			switch r.Method {
			case http.MethodGet:
				indexHandler.GET(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	authRouter := http.NewServeMux()
	handler.Handle("/auth/", http.StripPrefix("/auth", authRouter))
	authRouter.HandleFunc("/google", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			middleware.Guest(googleOauthHandler.GET, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	authRouter.HandleFunc("/totp", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			middleware.Guest(totpHandler.GET, w, r)
		case http.MethodPost:
			middleware.Guest(totpHandler.POST, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	authRouter.HandleFunc("/google/callback", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			middleware.Guest(googleOauthCallbackHandler.GET, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	authRouter.HandleFunc("/google/setup/{code}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			middleware.Guest(googleOauthSetupHandler.GET, w, r)
		case http.MethodPost:
			middleware.Guest(googleOauthSetupHandler.POST, w, r)
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
		}
	})

	signupRouter := http.NewServeMux()
	handler.Handle("/signup/", http.StripPrefix("/signup", signupRouter))

	signupRouter.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			middleware.Guest(signupHandler.GET, w, r)
		case http.MethodPost:
			middleware.Guest(signupHandler.POST, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	signupRouter.HandleFunc("/verify/{code}", func(w http.ResponseWriter, r *http.Request) {
		middleware.Guest(signupVerifyHandler.GET, w, r)
	})

	forgotPasswordRouter := http.NewServeMux()
	handler.Handle("/forgot-password/", http.StripPrefix("/forgot-password", forgotPasswordRouter))
	forgotPasswordRouter.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			middleware.Guest(forgotPasswordHandler.GET, w, r)
		case http.MethodPost:
			middleware.Guest(forgotPasswordHandler.POST, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	forgotPasswordRouter.HandleFunc("/verify/{code}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			middleware.Guest(forgotPasswordVerifyHandler.GET, w, r)
		case http.MethodPost:
			middleware.Guest(forgotPasswordVerifyHandler.POST, w, r)
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

	handler.HandleFunc("/user/totp/setup", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			middleware.Auth(userHandlerTotpSetup.GET, w, r)
		case http.MethodPost:
			middleware.Auth(userHandlerTotpSetup.POST, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Upload router
	uploadRouter := http.NewServeMux()
	handler.Handle("/upload/", http.StripPrefix("/upload", uploadRouter))

	uploadRouter.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			middleware.Auth(uploadHandler.GET, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	uploadRouter.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			middleware.Auth(uploadHandler.POST, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	uploadRouter.HandleFunc("/init", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			middleware.Auth(initialisation.POST, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Download router
	downloadRouter := http.NewServeMux()
	handler.Handle("/download/", http.StripPrefix("/download", downloadRouter))
	downloadRouter.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			middleware.Auth(downloadHandler.GET, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	downloadRouter.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
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
		http.ServeFile(w, r, "public/robots.txt")
	})

	handler.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "public/favicon.ico")
	})

	fileServer := http.FileServer(http.Dir("./public"))
	handler.Handle("/public/", http.StripPrefix("/public", fileServer))

	return handler
}
