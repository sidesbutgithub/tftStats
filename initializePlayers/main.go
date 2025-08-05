package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/redis/go-redis/v9"

	"github.com/sidesbutgithub/tftStats/initializePlayers/models"
)

func main() {

	rated := []string{"challenger", "grandmaster", "master"}
	ranks := []string{"DIAMOND", "EMERALD", "PLATINUM", "GOLD", "SILVER", "BRONZE", "IRON"}
	divisions := []string{"I", "II", "III", "IV"}
	apiKey := os.Getenv("RIOT_API_KEY")

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

	initialQueueLen, err := redisClient.LLen(context.Background(), "playersQueue").Result()
	if err != nil {
		log.Print(err)
		log.Fatal("Error getting queue len")
	}

	if initialQueueLen != 0 {
		log.Printf("Queue already contains %d entries, exiting initialization service", initialQueueLen)
		return
	}

	for _, ladder := range rated {
		res, err := http.Get(fmt.Sprintf("https://na1.api.riotgames.com/tft/league/v1/%s?queue=RANKED_TFT&api_key=%s", ladder, apiKey))
		if err != nil {
			log.Print(err)
			log.Printf("Error with http request for %s, skipping", ladder)
			continue
		}
		b, err := io.ReadAll(res.Body)
		if err != nil {
			log.Print(err)
			log.Printf("Error getting body data for %s, skipping", ladder)
			continue
		}

		defer res.Body.Close()

		var rankData models.RatedPlayersResponse
		err = json.Unmarshal(b, &rankData)
		if err != nil {
			log.Print(err)
			log.Printf("Error unmarshalling body data for %s, skipping", ladder)
			continue
		}

		if len(rankData.Entries) == 0 {
			log.Printf("No matches in %s, getting players for next rank", ladder)
			continue
		}

		initialPlayers := make([]string, 0)

		for _, data := range rankData.Entries {
			initialPlayers = append(initialPlayers, data.Puuid)
		}

		redisClient.RPush(context.Background(), "playersQueue", initialPlayers)
		log.Printf("initialized player queue with %d players from rank %s, initization service complete", len(initialPlayers), ladder)
		return
	}

	for _, rank := range ranks {
		for _, division := range divisions {
			res, err := http.Get(fmt.Sprintf("https://na1.api.riotgames.com/tft/league/v1/entries/%s/%s?queue=RANKED_TFT&page=1&api_key=%s", rank, division, apiKey))
			if err != nil {
				log.Print(err)
				log.Printf("Error with http request for %s %s, skipping", rank, division)
				continue
			}

			b, err := io.ReadAll(res.Body)
			if err != nil {
				log.Print(err)
				log.Printf("Error getting body data for %s %s, skipping", rank, division)
				continue
			}
			defer res.Body.Close()

			var rankData models.RankedPlayersResponse
			err = json.Unmarshal(b, &rankData)
			if err != nil {
				log.Print(err)
				log.Printf("Error unmarshalling body data for %s %s, skipping", rank, division)
				continue
			}

			if len(rankData) == 0 {
				log.Printf("No matches in %s %s, getting players for next rank", rank, division)
				continue
			}

			initialPlayers := make([]string, 0)

			for _, data := range rankData {
				initialPlayers = append(initialPlayers, data.Puuid)
			}

			redisClient.RPush(context.Background(), "playersQueue", initialPlayers)
			log.Printf("Initialized player queue with %d players from rank %s %s, initization service complete", len(initialPlayers), rank, division)
			return
		}
	}
}
