package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/sidesbutgithub/tftStats/matchCrawler/internal/database"
	"github.com/sidesbutgithub/tftStats/matchCrawler/internal/models"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Failed to load .env file")
	}

	//connect to postgresdb
	dbURI := os.Getenv("DB_URI")

	var pgDb database.PostgresDB

	err = pgDb.ConnectPostgres(dbURI)
	if err != nil {
		log.Fatal("Failed to connect to postgres")
	}

	riotApiKey := os.Getenv("RIOT_API_KEY")
	res, err := http.Get(fmt.Sprintf("https://americas.api.riotgames.com/tft/match/v1/matches/NA1_5322362987?api_key=%s", riotApiKey))
	if err != nil {
		log.Fatal("Failed to get data")
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal("Failed to read body")
	}

	var bodyData models.MatchResponse

	err = json.Unmarshal(b, &bodyData)
	if err != nil {
		log.Fatal("Failed to unmarshall body data")
	}

	for _, participant := range bodyData.Info.Participants {

		if err == redis.Nil {
			fmt.Println("participant not in queue")
		} else if err != nil {
			fmt.Println(err)
			log.Fatal("unkown error occured")
		}

		for _, unit := range participant.Units {
			_, err := pgDb.InsertUnit(unit.CharacterID, int16(unit.Tier), unit.ItemNames, int16(participant.Placement))
			if err != nil {
				fmt.Println(err)
				log.Fatal("error writing test data to database")
			}
		}
	}
}
