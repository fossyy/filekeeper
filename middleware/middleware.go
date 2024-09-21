package middleware

import (
	"context"
	"fmt"
	"github.com/fossyy/filekeeper/app"
	"net/http"
	"strings"

	errorHandler "github.com/fossyy/filekeeper/handler/error"
	"github.com/fossyy/filekeeper/session"
	"github.com/fossyy/filekeeper/utils"
)

type wrapper struct {
	http.ResponseWriter
	request    *http.Request
	statusCode int
}

func (w *wrapper) WriteHeader(code int) {
	w.statusCode = code

	if code == http.StatusNotFound {
		w.Header().Set("Content-Type", "text/html")
		w.ResponseWriter.WriteHeader(code)
		errorHandler.NotFound(w.ResponseWriter, w.request)
		return
	}

	if code == http.StatusInternalServerError {
		w.Header().Set("Content-Type", "text/html")
		w.ResponseWriter.WriteHeader(code)
		errorHandler.InternalServerError(w.ResponseWriter, w.request)
		return
	}
	w.ResponseWriter.WriteHeader(code)

	return
}

func Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("upgrade") == "websocket" {
			hijacker, ok := writer.(http.Hijacker)
			if !ok {
				http.Error(writer, "Hijacking not supported", http.StatusInternalServerError)
				return
			}
			hijack, _, err := hijacker.Hijack()
			if err != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
			defer hijack.Close()
			rw := NewHijackWriter(hijack)
			app.Server.Logger.Info(fmt.Sprintf("%s %s %s %v", utils.ClientIP(request), "WEBSOCKET", request.RequestURI, http.StatusSwitchingProtocols))
			next.ServeHTTP(rw, request)
			return
		}

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
		app.Server.Logger.Info(fmt.Sprintf("%s %s %s %v", utils.ClientIP(request), request.Method, request.RequestURI, wrappedWriter.statusCode))
	})
}

func Auth(next http.HandlerFunc, w http.ResponseWriter, r *http.Request) {
	status, user, sessionID := session.GetSession(r)

	switch status {
	case session.Authorized:
		ctx := context.WithValue(r.Context(), "user", user)
		req := r.WithContext(ctx)
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
	case session.Suspicious:
		userSession := session.Get(sessionID)
		err := userSession.Delete()
		if err != nil {
			app.Server.Logger.Error(err)
		}
		err = session.RemoveSessionInfo(user.Email, sessionID)
		if err != nil {
			app.Server.Logger.Error(err)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:   "Session",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})
		http.Redirect(w, r, "/signin?error=suspicious_session", http.StatusSeeOther)
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
