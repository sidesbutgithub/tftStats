package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"

	"github.com/sidesbutgithub/tftStats/initializePlayers/models"
)

func main() {
	ranks := []string{"DIAMOND", "EMERALD", "PLATINUM", "GOLD", "SILVER", "BRONZE", "IRON"}
	divisions := []string{"I", "II", "III", "IV"}
	apiKey := os.Getenv("RIOT_API_KEY")
	maxRetries, err := strconv.Atoi(os.Getenv("REQUEST_RETRIES"))
	if err != nil {
		log.Print(err)
		log.Print("Error parsing max retries var from environment, setting default of 3 retries")
		maxRetries = 3
	}

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

	for _, rank := range ranks {
		for _, division := range divisions {
			res, err := http.Get(fmt.Sprintf("https://na1.api.riotgames.com/tft/league/v1/entries/%s/%s?queue=RANKED_TFT&page=1&api_key=%s", rank, division, apiKey))
			currRetries := 0
			for err != nil {
				currRetries += 1
				if currRetries > maxRetries {
					log.Print(err)
					log.Printf("Error getting data for %s %s exceeding max retries %d, skipping...", rank, division, maxRetries)
					continue
				}
				log.Print(err)
				log.Printf("Error getting data for %s %s, retrying...", rank, division)
			}
			currRetries = 0
			b, err := io.ReadAll(res.Body)
			for err != nil {
				currRetries += 1
				if currRetries > maxRetries {
					log.Print(err)
					log.Printf("Error reading body data for %s %s exceeding max retries %d, skipping...", rank, division, maxRetries)
					continue
				}
				log.Print(err)
				log.Printf("Error getting body data for %s %s, retrying...", rank, division)
			}
			defer res.Body.Close()

			var rankData models.RankedPlayersResponse
			err = json.Unmarshal(b, &rankData)
			for err != nil {
				currRetries += 1
				if currRetries > maxRetries {
					log.Print(err)
					log.Printf("Error unmarshalling body data for %s %s exceeding max retries %d, skipping...", rank, division, maxRetries)
					continue
				}
				log.Print(err)
				log.Printf("Error unmarshalling body data for %s %s, retrying...", rank, division)
			}

			if len(rankData) == 0 {
				log.Printf("no matches in %s %s, getting players for next rank", rank, division)
				continue
			}

			initialPlayers := make([]string, 0)

			for _, data := range rankData {
				initialPlayers = append(initialPlayers, data.Puuid)
			}

			redisClient.RPush(context.Background(), "playersQueue", initialPlayers)
			log.Printf("initialized player queue with %d players from rank %s %s, initization service complete", len(initialPlayers), rank, division)
			return
		}
	}
}
