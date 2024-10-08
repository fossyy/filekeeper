package client

import (
	"github.com/fossyy/filekeeper/handler/auth/forgotPassword"
	"github.com/fossyy/filekeeper/handler/auth/forgotPassword/verify"
	googleOauthHandler "github.com/fossyy/filekeeper/handler/auth/google"
	googleOauthCallbackHandler "github.com/fossyy/filekeeper/handler/auth/google/callback"
	googleOauthSetupHandler "github.com/fossyy/filekeeper/handler/auth/google/setup"
	"github.com/fossyy/filekeeper/handler/auth/signin"
	"github.com/fossyy/filekeeper/handler/auth/signup"
	"github.com/fossyy/filekeeper/handler/auth/signup/verify"
	totpHandler "github.com/fossyy/filekeeper/handler/auth/totp"
	fileHandler "github.com/fossyy/filekeeper/handler/file"
	deleteHandler "github.com/fossyy/filekeeper/handler/file/delete"
	downloadHandler "github.com/fossyy/filekeeper/handler/file/download"
	queryHandler "github.com/fossyy/filekeeper/handler/file/query"
	renameFileHandler "github.com/fossyy/filekeeper/handler/file/rename"
	fileTableHandler "github.com/fossyy/filekeeper/handler/file/table"
	uploadHandler "github.com/fossyy/filekeeper/handler/file/upload"
	visibilityHandler "github.com/fossyy/filekeeper/handler/file/visibility"
	indexHandler "github.com/fossyy/filekeeper/handler/index"
	logoutHandler "github.com/fossyy/filekeeper/handler/logout"
	userHandler "github.com/fossyy/filekeeper/handler/user"
	userHandlerResetPassword "github.com/fossyy/filekeeper/handler/user/ResetPassword"
	userSessionTerminateHandler "github.com/fossyy/filekeeper/handler/user/session/terminate"
	userHandlerTotpSetup "github.com/fossyy/filekeeper/handler/user/totp"
	websocketHandler "github.com/fossyy/filekeeper/handler/websocket"
	"github.com/fossyy/filekeeper/middleware"
	"net/http"
)

func SetupRoutes() *http.ServeMux {
	handler := http.NewServeMux()
	handler.HandleFunc("GET /{$}", indexHandler.GET)
	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	authRoute := http.NewServeMux()
	handler.Handle("/auth/", http.StripPrefix("/auth", authRoute))
	authRoute.Handle("GET /google", middleware.Guest(googleOauthHandler.GET))
	authRoute.Handle("GET /totp", middleware.Guest(totpHandler.GET))
	authRoute.Handle("POST /totp", middleware.Guest(totpHandler.POST))
	authRoute.Handle("GET /google/callback", middleware.Guest(googleOauthCallbackHandler.GET))
	authRoute.Handle("GET /google/setup/{code}", middleware.Guest(googleOauthSetupHandler.GET))
	authRoute.Handle("POST /google/setup/{code}", middleware.Guest(googleOauthSetupHandler.POST))
	authRoute.Handle("GET /signin", middleware.Guest(signinHandler.GET))
	authRoute.Handle("POST /signin", middleware.Guest(signinHandler.POST))
	authRoute.Handle("GET /signup", middleware.Guest(signupHandler.GET))
	authRoute.Handle("POST /signup", middleware.Guest(signupHandler.POST))
	authRoute.Handle("GET /signup/verify/{code}", middleware.Guest(signupVerifyHandler.GET))
	authRoute.Handle("GET /forgot-password", middleware.Guest(forgotPasswordHandler.GET))
	authRoute.Handle("POST /forgot-password", middleware.Guest(forgotPasswordHandler.POST))
	authRoute.Handle("GET /forgot-password/verify/{code}", middleware.Guest(forgotPasswordVerifyHandler.GET))
	authRoute.Handle("POST /forgot-password/verify/{code}", middleware.Guest(forgotPasswordVerifyHandler.POST))

	userRoute := http.NewServeMux()
	handler.Handle("/user/", http.StripPrefix("/user", userRoute))
	userRoute.Handle("GET /{$}", middleware.Auth(userHandler.GET))
	userRoute.Handle("POST /reset-password", middleware.Auth(userHandlerResetPassword.POST))
	userRoute.Handle("DELETE /session/terminate/{id}", middleware.Auth(userSessionTerminateHandler.DELETE))
	userRoute.Handle("GET /totp/setup", middleware.Auth(userHandlerTotpSetup.GET))
	userRoute.Handle("POST /totp/setup", middleware.Auth(userHandlerTotpSetup.POST))

	handler.Handle("/ws", middleware.Auth(websocketHandler.GET))

	fileRoute := http.NewServeMux()
	handler.Handle("/file/", http.StripPrefix("/file", fileRoute))
	fileRoute.Handle("GET /{$}", middleware.Auth(fileHandler.GET))
	fileRoute.Handle("GET /table", middleware.Auth(fileTableHandler.GET))
	fileRoute.Handle("GET /query", middleware.Auth(queryHandler.GET))
	fileRoute.Handle("POST /{id}", middleware.Auth(uploadHandler.POST))
	fileRoute.Handle("DELETE /{id}", middleware.Auth(deleteHandler.DELETE))
	fileRoute.HandleFunc("GET /{id}", downloadHandler.GET)
	fileRoute.Handle("PUT /{id}", middleware.Auth(visibilityHandler.PUT))
	fileRoute.Handle("PATCH /{id}", middleware.Auth(renameFileHandler.PATCH))

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
