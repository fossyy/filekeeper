package cache

import (
	"fmt"
	"github.com/fossyy/filekeeper/db"
	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/utils"
	"github.com/google/uuid"
	"sync"
	"time"
)

type UserWithExpired struct {
	UserID   uuid.UUID
	Username string
	Email    string
	Password string
	AccessAt time.Time
	mu       sync.Mutex
}

var log *logger.AggregatedLogger
var userCache map[string]*UserWithExpired

func init() {
	log = logger.Logger()

	userCache = make(map[string]*UserWithExpired)
	ticker := time.NewTicker(time.Minute)

	go func() {
		for {
			<-ticker.C
			currentTime := time.Now()
			cacheClean := 0
			cleanID := utils.GenerateRandomString(10)
			log.Info(fmt.Sprintf("Cache cleanup [user] [%s] initiated at %02d:%02d:%02d", cleanID, currentTime.Hour(), currentTime.Minute(), currentTime.Second()))

			for _, user := range userCache {
				user.mu.Lock()
				if currentTime.Sub(user.AccessAt) > time.Hour*8 {
					delete(userCache, user.Email)
					cacheClean++
				}
				user.mu.Unlock()
			}

			log.Info(fmt.Sprintf("Cache cleanup [user] [%s] completed: %d entries removed. Finished at %s", cleanID, cacheClean, time.Since(currentTime)))
		}
	}()
}

func GetUser(email string) (*UserWithExpired, error) {
	if user, ok := userCache[email]; ok {
		return user, nil
	}

	userData, err := db.DB.GetUser(email)
	if err != nil {
		return nil, err
	}

	userCache[email] = &UserWithExpired{
		UserID:   userData.UserID,
		Username: userData.Username,
		Email:    userData.Email,
		Password: userData.Password,
		AccessAt: time.Now(),
	}

	return userCache[email], nil
}

func DeleteUser(email string) {
	userCache[email].mu.Lock()
	defer userCache[email].mu.Unlock()

	delete(userCache, email)
}
