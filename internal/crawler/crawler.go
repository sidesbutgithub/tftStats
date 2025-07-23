package crawler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/sidesbutgithub/tftStats/matchCrawler/internal/database"
	"github.com/sidesbutgithub/tftStats/matchCrawler/internal/models"
)

//store data locally before writing as bulk insert queries significantly faster

type Crawler struct {
	mu *sync.Mutex
	wg *sync.WaitGroup

	Queue      *database.RedisDB
	DB         *database.PostgresDB
	CurrData   []database.Unit
	RiotApiKey string
	NumWorkers int
}

// adds the puuid to the queue and marks them as visited
func (crawlerInst *Crawler) AddPlayer(puuid string) error {
	err := crawlerInst.Queue.EnqueuePlayer(puuid)
	if err != nil {
		log.Print("Error enquing player")
		return err
	}
	err = crawlerInst.Queue.MarkPlayerVisited(puuid)
	if err != nil {
		log.Print("Error marking player as visited")
		return err
	}
	return nil
}

// adds the match to the queue and marks them as visited
func (crawlerInst *Crawler) AddMatch(matchId string) error {
	err := crawlerInst.Queue.EnqueueMatch(matchId)
	if err != nil {
		log.Print("Error enquing match")
		return err
	}
	err = crawlerInst.Queue.MarkMatchVisited(matchId)
	if err != nil {
		log.Print("Error marking match as visited")
		return err
	}
	return nil
}

func (crawlerInst *Crawler) InitializaPlayerQueue(initialQueueLen int) error {
	crawlerInst.mu.Lock()
	defer crawlerInst.mu.Unlock()
	limitHits := 0
	var res *http.Response
	var err error
	for {
		res, err = http.Get(fmt.Sprintf("https://na1.api.riotgames.com/tft/league/v1/challenger?queue=RANKED_TFT&api_key=%s", crawlerInst.RiotApiKey))
		if err != nil {
			log.Print("Failed to get match response")
			return err
		}
		if res.StatusCode == 200 {
			break
		}
		if res.StatusCode != 429 {
			log.Print(res.StatusCode)
			return errors.New("unexpected http status code")
		}

		limitHits += 1
		if limitHits == 1 {
			log.Print("hit lower rate limit 20 reqs/s, sleeping 1s before retrying")
			time.Sleep(time.Second)
		} else if limitHits < 4 { //possibly set max retries constant
			log.Print("hit greater rate limit 100 reqs/2 mins, sleeping 2 mins before retrying")
			time.Sleep(2 * time.Minute)
		} else {
			log.Print("hit rate limit max number of times, no more retying, quiting program")
			return errors.New("rate limit exceeded excessively")
		}
	}

	var bodyData models.LeagueResponse

	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		log.Print("Failed to read response body")
		return err
	}

	err = json.Unmarshal(b, &bodyData)
	if err != nil {
		log.Print("Failed to unmarshall body data")
		return err
	}

	//TODO: need to check lower ranks if initialLen more than ppl in rank T-T, also go down ranks until all players initialized

	for i := 0; i < initialQueueLen; i++ {
		err = crawlerInst.AddPlayer(bodyData.Entries[i].Puuid)
		if err != nil {
			log.Print("Failed to unmarshall body data")
			return err
		}
	}
	return nil
}

// adds the data of a given match to the database and adds all the participants of that match
func (crawlerInst *Crawler) GetMatchData(matchID string) error {
	/*
		visited, err := crawlerInst.Queue.CheckMatchVisited(matchID)
		if err != nil {
			log.Print("Error checking if match in queue")
			return err
		}
		if visited {
			log.Printf("match %s already visited", matchID)
			return nil
		}
		err = crawlerInst.Queue.MarkMatchVisited(matchID)
		if err != nil {
			log.Print("Error marking match as Visited")
			return err
		}
	*/
	crawlerInst.mu.Lock()
	defer crawlerInst.mu.Unlock()
	limitHits := 0
	var res *http.Response
	var err error
	for {
		res, err = http.Get(fmt.Sprintf("https://americas.api.riotgames.com/tft/match/v1/matches/%s?api_key=%s", matchID, crawlerInst.RiotApiKey))
		if err != nil {
			log.Print("Failed to get match response")
			return err
		}
		if res.StatusCode == 200 {
			break
		}
		if res.StatusCode != 429 {
			log.Print(res.StatusCode)
			return errors.New("unexpected http status code")
		}

		limitHits += 1
		if limitHits == 1 {
			log.Print("hit lower rate limit 20 reqs/s, sleeping 1s before retrying")
			time.Sleep(time.Second)
		} else if limitHits < 4 { //possibly set max retries constant
			log.Print("hit greater rate limit 100 reqs/2 mins, sleeping 2 mins before retrying")
			time.Sleep(2 * time.Minute)
		} else {
			log.Print("hit rate limit max number of times, no more retying, quiting program")
			return errors.New("rate limit exceeded excessively")
		}

	}

	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		log.Print("Failed to read response body")
		return err
	}

	var bodyData models.MatchResponse

	err = json.Unmarshal(b, &bodyData)
	if err != nil {
		log.Print("Failed to unmarshall body data")
		return err
	}

	for _, participant := range bodyData.Info.Participants {
		visited, err := crawlerInst.Queue.CheckPlayerVisited(participant.Puuid)
		if err != nil {
			log.Print("Error checking if player was visited")
			return err
		}
		if !visited {
			err = crawlerInst.AddPlayer(participant.Puuid)
			if err != nil {
				log.Print("Failed to add player to queue and visited")
				return err
			}
		}
		for _, unit := range participant.Units {
			_, err := crawlerInst.DB.InsertUnit(unit.CharacterID, int16(unit.Tier), unit.ItemNames, int16(participant.Placement))
			if err != nil {
				fmt.Println("error adding unit to DB")
				return err
			}
		}
	}
	return nil
}

// inserts the last 20 matches of the given puuid into the matches queue and marks them as visited if not already visited
func (crawlerInst *Crawler) GetMatches(puuid string) error {
	/*
		visited, err := crawlerInst.Queue.CheckPlayerVisited(puuid)
		if err != nil {
			log.Print("Error checking if player has been visited")
			return err
		}
		if visited {
			log.Printf("player %s already visited", puuid)
			return nil
		}
		err = crawlerInst.Queue.MarkPlayerVisited(puuid)
		if err != nil {
			log.Print("Error marking player as Visited")
			return err
		}
	*/
	crawlerInst.mu.Lock()
	defer crawlerInst.mu.Unlock()
	limitHits := 0
	var res *http.Response
	var err error
	for {
		res, err = http.Get(fmt.Sprintf("https://americas.api.riotgames.com/tft/match/v1/matches/by-puuid/%s/ids?start=0&count=20&api_key=%s", puuid, crawlerInst.RiotApiKey))
		if err != nil {
			log.Print("Failed to get match response")
			return err
		}
		if res.StatusCode == 200 {
			break
		}
		if res.StatusCode != 429 {
			log.Print(res.StatusCode)
			return errors.New("unexpected http status code")
		}

		limitHits += 1
		if limitHits == 1 {
			log.Print("hit lower rate limit 20 reqs/s, sleeping 1s before retrying")
			time.Sleep(time.Second)
		} else if limitHits < 4 { //possibly set max retries constant
			log.Print("hit greater rate limit 100 reqs/2 mins, sleeping 2 mins before retrying")
			time.Sleep(2 * time.Minute)
		} else {
			log.Print("hit rate limit max number of times, no more retying, quiting program")
			return errors.New("rate limit exceeded excessively")
		}

	}

	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		log.Print("Failed to read response body")
		return err
	}

	var bodyData []string

	err = json.Unmarshal(b, &bodyData)
	if err != nil {
		log.Print("Failed to unmarshall body data")
		log.Print(b)
		log.Print(err)
		return err
	}

	for _, matchID := range bodyData {
		visited, err := crawlerInst.Queue.CheckMatchVisited(matchID)
		if err != nil {
			log.Print("Error checking if match was visited")
			return err
		}
		if !visited {
			err = crawlerInst.AddMatch(matchID)
			if err != nil {
				log.Print("Failed to add match to queue and visited")
				return err
			}
		}
	}
	return nil
}
