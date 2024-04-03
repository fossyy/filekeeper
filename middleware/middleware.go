package middleware

import (
	"errors"
	"fmt"
	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/session"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/utils"
	"net/http"
	"strings"
)

var log *logger.AggregatedLogger

func init() {
	log = logger.Logger()
}

func Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		address := strings.Split(utils.Getenv("CORS_LIST"), ",")

		for _, addr := range address {
			if request.Host == addr {
				writer.Header().Set("Access-Control-Allow-Origin", fmt.Sprintf("%s://%s", utils.Getenv("CORS_PROTO"), addr))
			}
		}
		writer.Header().Set("Access-Control-Allow-Methods", fmt.Sprintf("%s, OPTIONS", utils.Getenv("CORS_METHODS")))
		writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		next.ServeHTTP(writer, request)
		log.Info(fmt.Sprintf("%s %s %s \n", utils.ClientIP(request), request.Method, request.RequestURI))
	})
}

func Auth(next http.HandlerFunc, w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("Session")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			http.SetCookie(w, &http.Cookie{
				Name:  "redirect",
				Value: r.URL.String(),
				Path:  "/",
			})
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	storeSession, err := session.Store.Get(cookie.Value)
	if err != nil {
		if errors.Is(err, &session.SessionNotFound{}) {
			http.SetCookie(w, &http.Cookie{
				Name:   "Session",
				Value:  "",
				MaxAge: -1,
			})
		}
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	userSession := GetUser(storeSession)
	if userSession.Authenticated {
		next.ServeHTTP(w, r)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:  "redirect",
		Value: r.URL.String(),
		Path:  "/",
	})
	http.Redirect(w, r, "/signin", http.StatusSeeOther)
	return
}

func Guest(next http.HandlerFunc, w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("Session")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			next.ServeHTTP(w, r)
			return
		}
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	storeSession, err := session.Store.Get(cookie.Value)
	if err != nil {
		if errors.Is(err, &session.SessionNotFound{}) {
			http.SetCookie(w, &http.Cookie{
				Name:   "Session",
				Value:  "",
				MaxAge: -1,
			})
			next.ServeHTTP(w, r)
			return
		} else {
			log.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	userSession := GetUser(storeSession)
	if !userSession.Authenticated {
		next.ServeHTTP(w, r)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return
}

func GetUser(s *session.Session) types.User {
	val := s.Values["user"]
	var userSession = types.User{}
	userSession, ok := val.(types.User)
	if !ok {
		return types.User{Authenticated: false}
	}
	return userSession
}
