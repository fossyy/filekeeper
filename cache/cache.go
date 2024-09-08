package cache

import (
	"context"
	"fmt"
	"github.com/fossyy/filekeeper/types"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisServer struct {
	client   *redis.Client
	database types.Database
}

func NewRedisServer(host, port, password string, db types.Database) types.CachingServer {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       0,
	})
	return &RedisServer{client: client, database: db}
}

func (r *RedisServer) GetCache(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func (r *RedisServer) SetCache(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	err := r.client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisServer) DeleteCache(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisServer) GetKeys(ctx context.Context, pattern string) ([]string, error) {
	var cursor uint64
	var keys []string
	for {
		var newKeys []string
		var err error

		newKeys, cursor, err = r.client.Scan(ctx, cursor, pattern, 0).Result()
		if err != nil {
			return nil, err
		}

		keys = append(keys, newKeys...)

		if cursor == 0 {
			break
		}
	}
	return keys, nil
}
