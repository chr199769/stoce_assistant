package redis

import (
	"context"
	"time"

	"fmt"
	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

func Init() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // In a real app, this should come from config
		Password: "",               // no password set
		DB:       0,                // use default DB
	})
	
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		fmt.Printf("Warning: Failed to connect to Redis: %v. Caching will be disabled.\n", err)
		// We don't nil the rdb because the client handles reconnection, 
		// but requests will fail. We should handle errors in Get/Set gracefully.
	}
}

func Get(ctx context.Context, key string) (string, error) {
	if rdb == nil {
		return "", fmt.Errorf("redis not initialized")
	}
	return rdb.Get(ctx, key).Result()
}

func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if rdb == nil {
		return fmt.Errorf("redis not initialized")
	}
	return rdb.Set(ctx, key, value, expiration).Err()
}

func GetClient() *redis.Client {
	return rdb
}
