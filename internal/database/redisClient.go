package database

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

type RedisDB struct {
	Client  *redis.Client
	Context context.Context
}

func (db *RedisDB) ConnectRedis(redisHost string, redisPort string, redisPassword string, redisDBNum int) {
	db.Client = redis.NewClient(&redis.Options{
		Addr:     redisHost + ":" + redisPort,
		Password: redisPassword, // No password set
		DB:       redisDBNum,    // Use default DB
	})
	db.Context = context.Background()
	err := db.Client.Ping(db.Context).Err()
	if err != nil {
		log.Fatal("Failed to connect to Redis")
	}
}
