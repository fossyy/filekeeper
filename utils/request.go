package utils

import (
	"golang.org/x/crypto/bcrypt"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

func ClientIP(request *http.Request) string {
	ip := request.Header.Get("X-Real-IP")
	if ip == "" {
		ip = request.Header.Get("X-Forwarded-For")
		if ip == "" {
			ip = request.RemoteAddr
		}
	}

	if strings.Contains(ip, ",") {
		ips := strings.Split(ip, ",")
		ip = strings.TrimSpace(ips[0])
	}

	if strings.Contains(ip, ":") {
		ips := strings.Split(ip, ":")
		ip = ips[0]
	}

	return ip
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	var result strings.Builder
	for i := 0; i < length; i++ {
		randomIndex := seededRand.Intn(len(charset))
		result.WriteString(string(charset[randomIndex]))
	}
	return result.String()
}
