package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	ranks := []string{"DIAMOND", "EMERALD", "PLATINUM", "GOLD", "SILVER", "BRONZE", "IRON"}
	divisions := []string{"I", "II", "III", "IV"}
	apiKey := os.Getenv("RIOT_API_KEY")
	for _, rank := range ranks {
		for _, division := range divisions {
			res, err := http.Get(fmt.Sprintf("https://na1.api.riotgames.com/tft/league/v1/entries/%s/%s?queue=RANKED_TFT&page=1&api_key=%s", rank, division, apiKey))
			if err != nil {
				log.Print(err)
				log.Printf("Error getting data for %s %s, skipping...", rank, division)
				continue
			}
			defer res.Body.Close()
			b, err := io.ReadAll(res.Body)

			var rankData RankedPlayersResponse
			err = json.Unmarshal(b, &rankData)

		}
	}
}
