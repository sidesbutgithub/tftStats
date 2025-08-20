package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	opt, err := redis.ParseURL(os.Getenv("REDIS_URI"))
	if err != nil {
		log.Print(err)
		log.Print(os.Getenv("REDIS_URI"))
		log.Fatal("Error parsing Redis connection string")
	}

	redisClient := redis.NewClient(opt)
	err = redisClient.Ping(context.Background()).Err()
	if err != nil {
		log.Fatal("Failed to connect to Redis")
	}

	primaryTicker := time.NewTicker(time.Second)
	secondaryTicker := time.NewTicker(2 * time.Minute)
	done := make(chan bool)

	prevQueues, err := redisClient.Exists(context.Background(), "rateQueue1", "rateQueue2").Result()
	if err != nil {
		log.Print("failed to check pre-existing queues")
		log.Fatal(err)
	}
	if prevQueues == 0 { //if there were no previous records of rate limits, initialize the rate queues
		//do this in parallel cuz of 2 different queues
		go func() {
			calls := []string{"i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i"}
			redisClient.RPush(context.Background(), "rateQueue1", calls)
		}()

		go func() {
			calls := []string{"i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i"}
			redisClient.RPush(context.Background(), "rateQueue2", calls)
		}()
	}

	go func() {
		for {
			<-primaryTicker.C
			redisClient.Del(context.Background(), "rateQueue1")
			//array of 20 elements to represent available riot api calls
			calls := []string{"i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i"}
			redisClient.RPush(context.Background(), "rateQueue1", calls)
		}
	}()

	go func() {
		for {
			<-secondaryTicker.C
			redisClient.Del(context.Background(), "rateQueue2")
			//array of 100 elements to represent available riot api calls
			calls := []string{"i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i", "i"}
			redisClient.RPush(context.Background(), "rateQueue2", calls)

		}
	}()

	<-done
}
