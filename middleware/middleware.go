package middleware

import (
	"bufio"
	"context"
	"fmt"
	"net"
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
	if code == http.StatusNotFound {
		w.Header().Set("Content-Type", "text/html")
		errorHandler.ALL(w.ResponseWriter, w.request)
		return
	}
	w.ResponseWriter.WriteHeader(code)
	w.statusCode = code
	return
}

func Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Upgrade") == "websocket" {
			hijacker, ok := w.(http.Hijacker)
			if !ok {
				http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
				return
			}
			hijackConn, _, err := hijacker.Hijack()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer hijackConn.Close()
			rw := NewResponseWriter(hijackConn)
			next.ServeHTTP(rw, r)
			log.Info(fmt.Sprintf("%s %s %s \n", utils.ClientIP(r), "WEBSOCKET", r.RequestURI))
			return
		}

		address := strings.Split(utils.Getenv("CORS_LIST"), ",")

		for _, addr := range address {
			if r.Host == addr {
				w.Header().Set("Access-Control-Allow-Origin", fmt.Sprintf("%s://%s", utils.Getenv("CORS_PROTO"), addr))
			}
		}

		wrappedWriter := &wrapper{
			ResponseWriter: w,
			request:        r,
			statusCode:     http.StatusOK,
		}

		w.Header().Set("Access-Control-Allow-Methods", fmt.Sprintf("%s, OPTIONS", utils.Getenv("CORS_METHODS")))
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		next.ServeHTTP(wrappedWriter, r)
		log.Info(fmt.Sprintf("%s %s %s %v \n", utils.ClientIP(r), r.Method, r.RequestURI, wrappedWriter.statusCode))
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

func NewResponseWriter(conn net.Conn) http.ResponseWriter {
	return &responseWriter{
		conn: conn,
	}
}

type responseWriter struct {
	conn net.Conn
}

func (rw *responseWriter) Header() http.Header {
	return http.Header{}
}

func (rw *responseWriter) WriteHeader(statusCode int) {
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	return rw.conn.Write(data)
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return rw.conn, bufio.NewReadWriter(bufio.NewReader(rw.conn), bufio.NewWriter(rw.conn)), nil
}
