package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	errorHandler "github.com/fossyy/filekeeper/handler/error"
	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/session"
	"github.com/fossyy/filekeeper/utils"
)

var log *logger.AggregatedLogger

func init() {
	log = logger.Logger()
}

type wrapper struct {
	http.ResponseWriter
	request    *http.Request
	statusCode int
}

func (w *wrapper) WriteHeader(code int) {
	switch code {
	case http.StatusNotFound:
		w.Header().Set("Content-Type", "text/html")
		errorHandler.NotFound(w.ResponseWriter, w.request)
		return
	case http.StatusInternalServerError:
		w.Header().Set("Content-Type", "text/html")
		errorHandler.InternalServerError(w.ResponseWriter, w.request)
		return
	default:
		w.ResponseWriter.WriteHeader(code)
		w.statusCode = code
		return
	}
}

func Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		address := strings.Split(utils.Getenv("CORS_LIST"), ",")

		for _, addr := range address {
			if request.Host == addr {
				writer.Header().Set("Access-Control-Allow-Origin", fmt.Sprintf("%s://%s", utils.Getenv("CORS_PROTO"), addr))
			}
		}

		wrappedWriter := &wrapper{
			ResponseWriter: writer,
			request:        request,
			statusCode:     http.StatusOK,
		}

		writer.Header().Set("Access-Control-Allow-Methods", fmt.Sprintf("%s, OPTIONS", utils.Getenv("CORS_METHODS")))
		writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		next.ServeHTTP(wrappedWriter, request)
		log.Info(fmt.Sprintf("%s %s %s %v \n", utils.ClientIP(request), request.Method, request.RequestURI, wrappedWriter.statusCode))
	})
}

func Auth(next http.HandlerFunc, w http.ResponseWriter, r *http.Request) {
	status, user, sessionID := session.GetSession(r)

	switch status {
	case session.Authorized:
		userSession := session.GetSessionInfo(user.Email, sessionID)
		userSession.UpdateAccessTime()

		ctx := context.WithValue(r.Context(), "user", user)
		req := r.WithContext(ctx)
		r.Context().Value("user")
		next.ServeHTTP(w, req)
		return
	case session.Unauthorized:
		if r.RequestURI != "/logout" {
			http.SetCookie(w, &http.Cookie{
				Name:  "redirect",
				Value: r.RequestURI,
				Path:  "/",
			})
		}
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	case session.InvalidSession:
		http.SetCookie(w, &http.Cookie{
			Name:   "Session",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	default:
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
}

func Guest(next http.HandlerFunc, w http.ResponseWriter, r *http.Request) {
	status, _, _ := session.GetSession(r)

	switch status {
	case session.Authorized:
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	case session.Unauthorized:
		next.ServeHTTP(w, r)
		return
	case session.InvalidSession:
		http.SetCookie(w, &http.Cookie{
			Name:   "Session",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})
		next.ServeHTTP(w, r)
		return
	}
}
