package client

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
	userHandler "github.com/fossyy/filekeeper/handler/user"
	userHandlerResetPassword "github.com/fossyy/filekeeper/handler/user/ResetPassword"
	userSessionTerminateHandler "github.com/fossyy/filekeeper/handler/user/session/terminate"
	userHandlerTotpSetup "github.com/fossyy/filekeeper/handler/user/totp"
	"github.com/fossyy/filekeeper/middleware"
	"net/http"
)

func SetupRoutes() *http.ServeMux {
	handler := http.NewServeMux()

	handler.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		indexHandler.GET(w, r)
	})

	handler.HandleFunc("GET /auth/google", func(w http.ResponseWriter, r *http.Request) {
		middleware.Guest(googleOauthHandler.GET, w, r)
	})

	handler.HandleFunc("GET /auth/totp", func(w http.ResponseWriter, r *http.Request) {
		middleware.Guest(totpHandler.GET, w, r)
	})

	handler.HandleFunc("POST /auth/totp", func(w http.ResponseWriter, r *http.Request) {
		middleware.Guest(totpHandler.POST, w, r)
	})

	handler.HandleFunc("GET /auth/google/callback", func(w http.ResponseWriter, r *http.Request) {
		middleware.Guest(googleOauthCallbackHandler.GET, w, r)
	})

	handler.HandleFunc("GET /auth/google/setup/{code}", func(w http.ResponseWriter, r *http.Request) {
		middleware.Guest(googleOauthSetupHandler.GET, w, r)
	})
	handler.HandleFunc("POST /auth/google/setup/{code}", func(w http.ResponseWriter, r *http.Request) {
		middleware.Guest(googleOauthSetupHandler.POST, w, r)
	})

	handler.HandleFunc("GET /signin", func(w http.ResponseWriter, r *http.Request) {
		middleware.Guest(signinHandler.GET, w, r)
	})

	handler.HandleFunc("POST /signin", func(w http.ResponseWriter, r *http.Request) {
		middleware.Guest(signinHandler.POST, w, r)
	})

	handler.HandleFunc("GET /signup", func(w http.ResponseWriter, r *http.Request) {
		middleware.Guest(signupHandler.GET, w, r)
	})

	handler.HandleFunc("POST /signup", func(w http.ResponseWriter, r *http.Request) {
		middleware.Guest(signupHandler.POST, w, r)
	})

	handler.HandleFunc("GET /signup/verify/{code}", func(w http.ResponseWriter, r *http.Request) {
		middleware.Guest(signupVerifyHandler.GET, w, r)
	})

	handler.HandleFunc("GET /forgot-password", func(w http.ResponseWriter, r *http.Request) {
		middleware.Guest(forgotPasswordHandler.GET, w, r)
	})

	handler.HandleFunc("POST /forgot-password", func(w http.ResponseWriter, r *http.Request) {
		middleware.Guest(forgotPasswordHandler.POST, w, r)
	})

	handler.HandleFunc("GET /forgot-password/verify/{code}", func(w http.ResponseWriter, r *http.Request) {
		middleware.Guest(forgotPasswordVerifyHandler.GET, w, r)
	})

	handler.HandleFunc("POST /forgot-password/verify/{code}", func(w http.ResponseWriter, r *http.Request) {
		middleware.Guest(forgotPasswordVerifyHandler.POST, w, r)
	})

	handler.HandleFunc("GET /user", func(w http.ResponseWriter, r *http.Request) {
		middleware.Auth(userHandler.GET, w, r)
	})

	handler.HandleFunc("POST /user/reset-password", func(w http.ResponseWriter, r *http.Request) {
		middleware.Auth(userHandlerResetPassword.POST, w, r)
	})

	handler.HandleFunc("DELETE /user/session/terminate/{id}", func(w http.ResponseWriter, r *http.Request) {
		middleware.Auth(userSessionTerminateHandler.DELETE, w, r)
	})

	handler.HandleFunc("GET /user/totp/setup", func(w http.ResponseWriter, r *http.Request) {
		middleware.Auth(userHandlerTotpSetup.GET, w, r)
	})

	handler.HandleFunc("POST /user/totp/setup", func(w http.ResponseWriter, r *http.Request) {
		middleware.Auth(userHandlerTotpSetup.POST, w, r)
	})

	handler.HandleFunc("GET /upload", func(w http.ResponseWriter, r *http.Request) {
		middleware.Auth(uploadHandler.GET, w, r)
	})

	handler.HandleFunc("POST /upload/{id}", func(w http.ResponseWriter, r *http.Request) {
		middleware.Auth(uploadHandler.POST, w, r)
	})

	handler.HandleFunc("GET /download", func(w http.ResponseWriter, r *http.Request) {
		middleware.Auth(downloadHandler.GET, w, r)
	})

	handler.HandleFunc("GET /download/{id}", func(w http.ResponseWriter, r *http.Request) {
		downloadFileHandler.GET(w, r)
	})

	handler.HandleFunc("GET /logout", func(w http.ResponseWriter, r *http.Request) {
		middleware.Auth(logoutHandler.GET, w, r)
	})

	handler.HandleFunc("GET /robots.txt", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "public/robots.txt")
	})

	handler.HandleFunc("GET /sitemap.xml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "public/sitemap.xml")
	})

	handler.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "public/favicon.ico")
	})

	fileServer := http.FileServer(http.Dir("./public"))
	handler.Handle("/public/", http.StripPrefix("/public", fileServer))

	return handler
}
