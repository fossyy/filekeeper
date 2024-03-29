package utils

import (
	"fmt"
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

func ConvertFileSize(byte int) string {
	if byte < 1024 {
		return fmt.Sprintf("%d B", byte)
	} else if byte < 1024*1024 {
		return fmt.Sprintf("%d KB", byte/1024)
	} else if byte < 1024*1024*1024 {
		return fmt.Sprintf("%d MB", byte/(1024*1024))
	} else {
		return fmt.Sprintf("%d GB", byte/(1024*1024*1024))
	}
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
