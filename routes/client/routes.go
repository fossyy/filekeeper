package client

import (
	googleOauthHandler "github.com/fossyy/filekeeper/handler/auth/google"
	googleOauthCallbackHandler "github.com/fossyy/filekeeper/handler/auth/google/callback"
	googleOauthSetupHandler "github.com/fossyy/filekeeper/handler/auth/google/setup"
	totpHandler "github.com/fossyy/filekeeper/handler/auth/totp"
	fileHandler "github.com/fossyy/filekeeper/handler/file"
	deleteHandler "github.com/fossyy/filekeeper/handler/file/delete"
	downloadHandler "github.com/fossyy/filekeeper/handler/file/download"
	queryHandler "github.com/fossyy/filekeeper/handler/file/query"
	renameFileHandler "github.com/fossyy/filekeeper/handler/file/rename"
	fileTableHandler "github.com/fossyy/filekeeper/handler/file/table"
	uploadHandler "github.com/fossyy/filekeeper/handler/file/upload"
	visibilityHandler "github.com/fossyy/filekeeper/handler/file/visibility"
	forgotPasswordHandler "github.com/fossyy/filekeeper/handler/forgotPassword"
	forgotPasswordVerifyHandler "github.com/fossyy/filekeeper/handler/forgotPassword/verify"
	indexHandler "github.com/fossyy/filekeeper/handler/index"
	logoutHandler "github.com/fossyy/filekeeper/handler/logout"
	signinHandler "github.com/fossyy/filekeeper/handler/signin"
	signupHandler "github.com/fossyy/filekeeper/handler/signup"
	signupVerifyHandler "github.com/fossyy/filekeeper/handler/signup/verify"
	userHandler "github.com/fossyy/filekeeper/handler/user"
	userHandlerResetPassword "github.com/fossyy/filekeeper/handler/user/ResetPassword"
	userSessionTerminateHandler "github.com/fossyy/filekeeper/handler/user/session/terminate"
	userHandlerTotpSetup "github.com/fossyy/filekeeper/handler/user/totp"
	"github.com/fossyy/filekeeper/middleware"
	"net/http"
)

func SetupRoutes() *http.ServeMux {
	handler := http.NewServeMux()

	// Index
	handler.HandleFunc("GET /{$}", indexHandler.GET)

	// Auth Routes
	handler.Handle("GET /auth/google", middleware.Guest(googleOauthHandler.GET))
	handler.Handle("GET /auth/totp", middleware.Guest(totpHandler.GET))
	handler.Handle("POST /auth/totp", middleware.Guest(totpHandler.POST))
	handler.Handle("GET /auth/google/callback", middleware.Guest(googleOauthCallbackHandler.GET))
	handler.Handle("GET /auth/google/setup/{code}", middleware.Guest(googleOauthSetupHandler.GET))
	handler.Handle("POST /auth/google/setup/{code}", middleware.Guest(googleOauthSetupHandler.POST))

	// Signin/Signup/Forgot Password
	handler.Handle("GET /signin", middleware.Guest(signinHandler.GET))
	handler.Handle("POST /signin", middleware.Guest(signinHandler.POST))
	handler.Handle("GET /signup", middleware.Guest(signupHandler.GET))
	handler.Handle("POST /signup", middleware.Guest(signupHandler.POST))
	handler.Handle("GET /signup/verify/{code}", middleware.Guest(signupVerifyHandler.GET))
	handler.Handle("GET /forgot-password", middleware.Guest(forgotPasswordHandler.GET))
	handler.Handle("POST /forgot-password", middleware.Guest(forgotPasswordHandler.POST))
	handler.Handle("GET /forgot-password/verify/{code}", middleware.Guest(forgotPasswordVerifyHandler.GET))
	handler.Handle("POST /forgot-password/verify/{code}", middleware.Guest(forgotPasswordVerifyHandler.POST))

	// User Routes
	handler.Handle("GET /user", middleware.Auth(userHandler.GET))
	handler.Handle("POST /user/reset-password", middleware.Auth(userHandlerResetPassword.POST))
	handler.Handle("DELETE /user/session/terminate/{id}", middleware.Auth(userSessionTerminateHandler.DELETE))
	handler.Handle("GET /user/totp/setup", middleware.Auth(userHandlerTotpSetup.GET))
	handler.Handle("POST /user/totp/setup", middleware.Auth(userHandlerTotpSetup.POST))

	// File Routes
	handler.Handle("GET /file", middleware.Auth(fileHandler.GET))
	handler.Handle("GET /file/table", middleware.Auth(fileTableHandler.GET))
	handler.Handle("GET /file/query", middleware.Auth(queryHandler.GET))
	handler.Handle("POST /file/{id}", middleware.Auth(uploadHandler.POST))
	handler.Handle("DELETE /file/{id}", middleware.Auth(deleteHandler.DELETE))
	handler.HandleFunc("GET /file/{id}", downloadHandler.GET)
	handler.Handle("PUT /file/{id}", middleware.Auth(visibilityHandler.PUT))
	handler.Handle("PATCH /file/{id}", middleware.Auth(renameFileHandler.PATCH))

	handler.Handle("GET /logout", middleware.Auth(logoutHandler.GET))

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
