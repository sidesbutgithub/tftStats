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
)

type LeagueResponse struct {
	Tier     string `json:"tier"`
	LeagueID string `json:"leagueId"`
	Queue    string `json:"queue"`
	Name     string `json:"name"`
	Entries  []struct {
		Puuid        string `json:"puuid"`
		LeaguePoints int    `json:"leaguePoints"`
		Rank         string `json:"rank"`
		Wins         int    `json:"wins"`
		Losses       int    `json:"losses"`
		Veteran      bool   `json:"veteran"`
		Inactive     bool   `json:"inactive"`
		FreshBlood   bool   `json:"freshBlood"`
		HotStreak    bool   `json:"hotStreak"`
	} `json:"entries"`
}

type MatchResponse struct {
	Metadata struct {
		DataVersion  string   `json:"data_version"`
		MatchID      string   `json:"match_id"`
		Participants []string `json:"participants"`
	} `json:"metadata"`
	Info struct {
		EndOfGameResult string  `json:"endOfGameResult"`
		GameCreation    int64   `json:"gameCreation"`
		GameID          int64   `json:"gameId"`
		GameDatetime    int64   `json:"game_datetime"`
		GameLength      float64 `json:"game_length"`
		GameVersion     string  `json:"game_version"`
		MapID           int     `json:"mapId"`
		Participants    []struct {
			Companion struct {
				ContentID string `json:"content_ID"`
				ItemID    int    `json:"item_ID"`
				SkinID    int    `json:"skin_ID"`
				Species   string `json:"species"`
			} `json:"companion"`
			GoldLeft  int `json:"gold_left"`
			LastRound int `json:"last_round"`
			Level     int `json:"level"`
			Missions  struct {
				PlayerScore2 int `json:"PlayerScore2"`
			} `json:"missions"`
			Placement            int     `json:"placement"`
			PlayersEliminated    int     `json:"players_eliminated"`
			Puuid                string  `json:"puuid"`
			RiotIDGameName       string  `json:"riotIdGameName"`
			RiotIDTagline        string  `json:"riotIdTagline"`
			TimeEliminated       float64 `json:"time_eliminated"`
			TotalDamageToPlayers int     `json:"total_damage_to_players"`
			Traits               []struct {
				Name        string `json:"name"`
				NumUnits    int    `json:"num_units"`
				Style       int    `json:"style"`
				TierCurrent int    `json:"tier_current"`
				TierTotal   int    `json:"tier_total"`
			} `json:"traits"`
			Units []struct {
				CharacterID string   `json:"character_id"`
				ItemNames   []string `json:"itemNames"`
				Name        string   `json:"name"`
				Rarity      int      `json:"rarity"`
				Tier        int      `json:"tier"`
			} `json:"units"`
			Win bool `json:"win"`
		} `json:"participants"`
		QueueID        int    `json:"queueId"`
		QueueID0       int    `json:"queue_id"`
		TftGameType    string `json:"tft_game_type"`
		TftSetCoreName string `json:"tft_set_core_name"`
		TftSetNumber   int    `json:"tft_set_number"`
	} `json:"info"`
}

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

	var bodyData MatchResponse

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
