package user

import (
	"fmt"
	"github.com/fossyy/filekeeper/db"
	"github.com/fossyy/filekeeper/logger"
	"github.com/fossyy/filekeeper/types"
	"sync"
	"time"
)

type Cache struct {
	users map[string]*types.UserWithExpired
	mu    sync.Mutex
}

var log *logger.AggregatedLogger
var UserCache *Cache

func init() {
	log = logger.Logger()

	UserCache = &Cache{users: make(map[string]*types.UserWithExpired)}
	ticker := time.NewTicker(time.Hour * 8)

	go func() {
		for {
			<-ticker.C
			currentTime := time.Now()
			cacheClean := 0
			log.Info(fmt.Sprintf("Cache cleanup initiated at %02d:%02d:%02d", currentTime.Hour(), currentTime.Minute(), currentTime.Second()))

			UserCache.mu.Lock()
			for _, user := range UserCache.users {
				if currentTime.Sub(user.AccessAt) > time.Hour*8 {
					delete(UserCache.users, user.Email)
					cacheClean++
				}
			}
			UserCache.mu.Unlock()

			log.Info(fmt.Sprintf("Cache cleanup completed: %d entries removed. Finished at %s", cacheClean, time.Since(currentTime)))
		}
	}()
}

func Get(email string) (*types.UserWithExpired, error) {
	UserCache.mu.Lock()
	defer UserCache.mu.Unlock()

	if user, ok := UserCache.users[email]; ok {
		return user, nil
	}

	var userData types.UserWithExpired
	err := db.DB.Table("users").Where("email = ?", email).First(&userData).Error
	if err != nil {
		return nil, err
	}

	UserCache.users[email] = &types.UserWithExpired{
		UserID:   userData.UserID,
		Username: userData.Username,
		Email:    userData.Email,
		Password: userData.Password,
		AccessAt: time.Now(),
	}

	return &userData, nil
}

func DeleteCache(email string) {
	UserCache.mu.Lock()
	defer UserCache.mu.Unlock()

	delete(UserCache.users, email)
}
