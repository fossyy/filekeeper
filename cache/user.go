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
}

type UserCache struct {
	users map[string]*UserWithExpired
	mu    sync.Mutex
}

var log *logger.AggregatedLogger
var userCache *UserCache

func init() {
	log = logger.Logger()

	userCache = &UserCache{users: make(map[string]*UserWithExpired)}
	ticker := time.NewTicker(time.Hour * 8)

	go func() {
		for {
			<-ticker.C
			currentTime := time.Now()
			cacheClean := 0
			cleanID := utils.GenerateRandomString(10)
			log.Info(fmt.Sprintf("Cache cleanup [user] [%s] initiated at %02d:%02d:%02d", cleanID, currentTime.Hour(), currentTime.Minute(), currentTime.Second()))

			userCache.mu.Lock()
			for _, user := range userCache.users {
				if currentTime.Sub(user.AccessAt) > time.Hour*8 {
					DeleteUser(user.Email)
					cacheClean++
				}
			}
			userCache.mu.Unlock()

			log.Info(fmt.Sprintf("Cache cleanup [user] [%s] completed: %d entries removed. Finished at %s", cleanID, cacheClean, time.Since(currentTime)))
		}
	}()
}

func GetUser(email string) (*UserWithExpired, error) {
	userCache.mu.Lock()
	defer userCache.mu.Unlock()

	if user, ok := userCache.users[email]; ok {
		return user, nil
	}

	userData, err := db.DB.GetUser(email)
	if err != nil {
		return nil, err
	}

	userCache.users[email] = &UserWithExpired{
		UserID:   userData.UserID,
		Username: userData.Username,
		Email:    userData.Email,
		Password: userData.Password,
		AccessAt: time.Now(),
	}

	return userCache.users[email], nil
}

func DeleteUser(email string) {
	userCache.mu.Lock()
	defer userCache.mu.Unlock()

	delete(userCache.users, email)
}
