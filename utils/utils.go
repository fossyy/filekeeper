package utils

import (
	"fmt"
	"github.com/fossyy/filekeeper/logger"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Env struct {
	value map[string]string
	mu    sync.Mutex
}

var env *Env
var log *logger.AggregatedLogger

func init() {
	env = &Env{value: map[string]string{}}
}

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

func Getenv(key string) string {
	env.mu.Lock()
	defer env.mu.Unlock()

	if val, ok := env.value[key]; ok {
		return val
	}

	err := godotenv.Load(".env")
	if err != nil {
		log.Error("Error loading .env file: %s", err)
	}

	val := os.Getenv(key)
	env.value[key] = val

	return val
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

func SanitizeFilename(filename string) string {
	invalidChars := []string{"\\", "/", ":", "*", "?", "\"", "<", ">", "|"}

	for _, char := range invalidChars {
		filename = strings.ReplaceAll(filename, char, "_")
	}

	return filename
}
