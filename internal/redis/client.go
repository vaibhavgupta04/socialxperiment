package redis

import (
	"context"
	"log"
	"github.com/redis/go-redis/v9"
	"github.com/gopro/internal/config"
)

var ctx = context.Background()

// InitRedis initializes a Redis client using the provided config.
func InitRedis(cfg *config.Config) *redis.Client {
	log.Print(cfg.RedisAddr, " ", cfg.RedisUser, " ", cfg.RedisPassword)
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
		Username: cfg.RedisUser,     // Optional, for ACL users
		Password: cfg.RedisPassword, // Required if set in Redis server
		DB:       0,                 // Default DB
	})

	// Test connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Redis connection failed: %v", err)
	}

	log.Println("Connected to Redis:", cfg.RedisAddr)
	return rdb
}
