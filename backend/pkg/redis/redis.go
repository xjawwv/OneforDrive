package redis

import (
	"context"
	"log"
	"os"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

var Client *goredis.Client

func InitRedis() {
	Client = goredis.NewClient(&goredis.Options{
		Addr:     getEnv("REDIS_HOST", "localhost") + ":" + getEnv("REDIS_PORT", "6379"),
		Password: "",
		DB:       0,
	})

	ctx := context.Background()
	for i := 0; i < 15; i++ {
		if err := Client.Ping(ctx).Err(); err == nil {
			log.Println("Connected to Redis")
			return
		}
		log.Printf("Waiting for Redis... attempt %d/15", i+1)
		time.Sleep(1 * time.Second)
	}
	log.Println("Warning: Redis not available, continuing without cache")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
