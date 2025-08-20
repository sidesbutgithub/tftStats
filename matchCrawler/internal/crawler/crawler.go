package crawler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"sync"

	"github.com/sidesbutgithub/tftStats/matchCrawler/internal/database"
	"github.com/sidesbutgithub/tftStats/matchCrawler/internal/databaseClients"
	"github.com/sidesbutgithub/tftStats/matchCrawler/internal/models"
	"github.com/sidesbutgithub/tftStats/matchCrawler/internal/utils"
	"golang.org/x/time/rate"
)

//store data locally before writing as bulk insert queries significantly faster

type Crawler struct {
	Mu *sync.Mutex
	Wg *sync.WaitGroup
	Rl *rate.Limiter

	Rdb              *databaseClients.RedisDB
	CurrData         []database.BulkInsertUnitsParams
	RiotApiKey       string
	MatchesStartTime string
	MatchWorkers     int
	PlayerWorkers    int
	MaxRetries       int
}

func (crawlerInst *Crawler) AddMatchIfNotVisited(matchId string) (bool, error) {
	crawlerInst.Mu.Lock()
	defer crawlerInst.Mu.Unlock()

	visited, err := crawlerInst.Rdb.CheckMatchVisited(matchId)
	if err != nil {
		log.Print("Error checking if match was visited")
		return false, err
	}
	if visited {
		return false, nil
	}
	err = crawlerInst.Rdb.EnqueueMatch(matchId)
	if err != nil {
		log.Print("Error enquing match")
		return false, err
	}
	err = crawlerInst.Rdb.MarkMatchVisited(matchId)
	if err != nil {
		log.Print("Error marking match as visited")
		return false, err
	}
	return true, nil
}

func (crawlerInst *Crawler) AddPlayerIfNotVisited(puuid string) (bool, error) {
	crawlerInst.Mu.Lock()
	defer crawlerInst.Mu.Unlock()

	visited, err := crawlerInst.Rdb.CheckPlayerVisited(puuid)
	if err != nil {
		log.Print("Error checking if player was visited")
		return false, err
	}
	if visited {
		return false, nil
	}
	err = crawlerInst.Rdb.EnqueuePlayer(puuid)
	if err != nil {
		log.Print("Error enquing player")
		return false, err
	}
	err = crawlerInst.Rdb.MarkPlayerVisited(puuid)
	if err != nil {
		log.Print("Error marking player as visited")
		return false, err
	}
	return true, nil
}

// adds the data of a given match to the database and adds all the participants of that match
func (crawlerInst *Crawler) GetMatchDataFromMatchID(matchID string) {
	reqAddress := fmt.Sprintf("https://americas.api.riotgames.com/tft/match/v1/matches/%s?api_key=%s", matchID, crawlerInst.RiotApiKey)

	_, err := crawlerInst.Rdb.Client.BLPop(context.Background(), 0, "rateQueue2").Result()
	if err != nil {
		log.Print("failed to wait for rate limit 2")
		log.Print(err)
		return
	}
	_, err = crawlerInst.Rdb.Client.BLPop(context.Background(), 0, "rateQueue1").Result()
	if err != nil {
		log.Print("failed to wait for rate limit 1")
		log.Print(err)
		return
	}

	b, err := utils.HandleHttpGetReqWithRetries(reqAddress, crawlerInst.MaxRetries)
	if err != nil {
		log.Print("Failed to get match, skipping")
		log.Print(err)
		return
	}
	if b == nil {
		log.Print("skipping queue item")
		return
	}

	var bodyData models.MatchResponse

	err = json.Unmarshal(b, &bodyData)
	if err != nil {
		log.Print("Failed to unmarshall body data, skipping")
		log.Print(reqAddress)
		log.Print(err)
		return
	}

	for _, participant := range bodyData.Info.Participants {
		_, err := crawlerInst.AddPlayerIfNotVisited(participant.Puuid)
		if err != nil {
			log.Print("error adding player to queue and visited set")
			log.Fatal(err)
		}
		crawlerInst.Mu.Lock()
		for _, unit := range participant.Units {
			//insert to slice within object to bulk write later
			if slices.Contains(unit.ItemNames, "TFT_Item_ThiefsGloves") {
				unit.ItemNames = []string{"TFT_Item_ThiefsGloves"}
			} else if slices.Contains(unit.ItemNames, "TFT5_Item_ThiefsGlovesRadiant") {
				unit.ItemNames = []string{"TFT5_Item_ThiefsGlovesRadiant"}
			}
			slices.Sort(unit.ItemNames)
			crawlerInst.CurrData = append(crawlerInst.CurrData, database.BulkInsertUnitsParams{
				Unitname:  unit.CharacterID,
				Starlevel: int16(unit.Tier),
				Items:     unit.ItemNames,
				Placement: int16(participant.Placement),
			})
		}
		crawlerInst.Mu.Unlock()
	}

}

// inserts the last 20 matches of the given puuid into the matches queue and marks them as visited if not already visited
func (crawlerInst *Crawler) GetMatchesFromPuuid(puuid string) {
	reqAddress := fmt.Sprintf("https://americas.api.riotgames.com/tft/match/v1/matches/by-puuid/%s/ids?start=0&start=0&startTime=%s&count=20&api_key=%s", puuid, crawlerInst.MatchesStartTime, crawlerInst.RiotApiKey)

	_, err := crawlerInst.Rdb.Client.BLPop(context.Background(), 0, "rateQueue2").Result()
	if err != nil {
		log.Print("failed to wait for rate limit 2")
		log.Fatal(err)
		return
	}
	_, err = crawlerInst.Rdb.Client.BLPop(context.Background(), 0, "rateQueue1").Result()
	if err != nil {
		log.Print("failed to wait for rate limit 1")
		log.Print(err)
		return
	}

	b, err := utils.HandleHttpGetReqWithRetries(reqAddress, crawlerInst.MaxRetries)
	if err != nil {
		log.Print("Failed to get match, skipping")
		log.Print(err)
		return
	}
	if b == nil {
		log.Print("skipping queue item")
		return
	}

	var bodyData []string

	err = json.Unmarshal(b, &bodyData)
	if err != nil {
		log.Print("Failed to unmarshall body data, skipping")
		log.Print(reqAddress)
		log.Print(err)
		return
	}

	for _, matchId := range bodyData {
		_, err := crawlerInst.AddMatchIfNotVisited(matchId)
		if err != nil {
			log.Print(err)
			log.Fatal("error adding player to queue and visited set")
		}
	}
}
