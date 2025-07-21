package main

import ("fmt"
		"net/http"
		"os"
		"log"
		"io"
		"github.com/joho/godotenv"
		"encoding/json"
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



func main(){
	err := godotenv.Load()
	if err != nil{
		log.Fatal("Failed to load .env file")
	}

	riotApiKey := os.Getenv("RIOT_API_KEY")
	res, err := http.Get(fmt.Sprintf("https://na1.api.riotgames.com/tft/league/v1/challenger?queue=RANKED_TFT&api_key=%s", riotApiKey))
	if err != nil{
		log.Fatal("Failed to get data")
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil{
		log.Fatal("Failed to read body")
	}

	var bodyData LeagueResponse

	err = json.Unmarshal(b, &bodyData)
	if err != nil{
		log.Fatal("Failed to unmarshall body data")
	}

	playerUUIDs := make([]string, 0, len(bodyData.Entries))

	for _, player := range bodyData.Entries {
		playerUUIDs = append(playerUUIDs, player.Puuid)
	}

	fmt.Println(len(playerUUIDs))
}