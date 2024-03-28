package middleware

import (
	"encoding/gob"
	"fmt"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/utils"
	"github.com/gorilla/sessions"
	"net/http"
)

var Store *sessions.CookieStore

func init() {
	authKeyOne := []byte{230, 131, 74, 255, 62, 51, 213, 168, 242, 70, 226, 115, 188, 243, 116, 226, 49, 12, 53, 17, 122, 162, 44, 185, 83, 53, 239, 16, 238, 154, 247, 222, 114, 86, 118, 242, 172, 97, 98, 47, 53, 219, 121, 89, 73, 124, 149, 116, 37, 122, 221, 47, 117, 142, 143, 139, 225, 180, 130, 93, 48, 83, 49, 165}
	encryptionKeyOne := []byte{20, 132, 251, 14, 203, 105, 189, 187, 10, 192, 68, 1, 100, 168, 213, 75, 127, 206, 42, 151, 208, 194, 38, 15, 34, 170, 28, 28, 55, 204, 45, 76}

	Store = sessions.NewCookieStore(
		authKeyOne,
		encryptionKeyOne,
	)

	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   15 * 60,
		HttpOnly: true,
	}

	gob.Register(types.User{})
}

func Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:8000")
		writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		next.ServeHTTP(writer, request)
		fmt.Printf("%s %s %s \n", utils.ClientIP(request), request.Method, request.RequestURI)
	})
}

func Auth(next http.HandlerFunc, w http.ResponseWriter, r *http.Request) {
	session, _ := Store.Get(r, "session")
	userSession := GetUser(session)
	if userSession.Authenticated {
		next.ServeHTTP(w, r)
		return
	}
	http.Redirect(w, r, "/signin", http.StatusSeeOther)
	return
}

func Guest(next http.HandlerFunc, w http.ResponseWriter, r *http.Request) {
	session, _ := Store.Get(r, "session")
	userSession := GetUser(session)
	if !userSession.Authenticated {
		next.ServeHTTP(w, r)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return
}

func GetUser(s *sessions.Session) types.User {
	val := s.Values["user"]
	var userSession = types.User{}
	userSession, ok := val.(types.User)
	if !ok {
		return types.User{Authenticated: false}
	}
	return userSession
}
