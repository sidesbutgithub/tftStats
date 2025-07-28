package main

import (
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
	"github.com/sidesbutgithub/tftStats/matchCrawler/internal/crawler"
	"github.com/sidesbutgithub/tftStats/matchCrawler/internal/database"
	"github.com/sidesbutgithub/tftStats/matchCrawler/internal/databaseClients"
	"golang.org/x/time/rate"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Failed to load .env file")
	}

	//connect to postgres
	dbURI := os.Getenv("DB_URI")
	var Pgdb databaseClients.PostgresDB
	defer Pgdb.CloseConn()
	err = Pgdb.ConnectPostgres(dbURI)
	if err != nil {
		log.Fatal("Failed to connect to Postgres DB")
	}

	//connect to redis
	var Rdb databaseClients.RedisDB
	RdbHost, RdbPort, RdbPW := os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT"), os.Getenv("REDIS_PASSWORD")
	RdbDbNum, err := strconv.Atoi(os.Getenv("REDIS_DB_NUM"))
	if err != nil {
		log.Fatal("unable to parse rdb num")
	}

	Rdb.ConnectRedis(RdbHost, RdbPort, RdbPW, RdbDbNum)
	if err != nil {
		log.Print(RdbHost, RdbPort, RdbPW, RdbDbNum)
		log.Fatal("Failed to connect to Redis DB")
	}

	riotApiKey := os.Getenv("RIOT_API_KEY")

	matchCrawler := &crawler.Crawler{
		Mu:         &sync.Mutex{},
		Wg:         &sync.WaitGroup{},
		Rl1:        rate.NewLimiter(20, 20),
		Rl2:        rate.NewLimiter(rate.Limit(float64(100)/float64(120)), 1),
		Rdb:        &Rdb,
		CurrData:   make([]database.BulkInsertUnitsParams, 0),
		RiotApiKey: riotApiKey,
		NumWorkers: 10,
	}
	//first run of crawler on each container will be same puuid
	startingPuuid := os.Getenv("STARTING_PUUID")
	matchCrawler.AddPlayerIfNotVisited(startingPuuid)
	maxDepth := 3
	currDepth := 0
	//main loop
	for {
		if currDepth >= maxDepth {
			log.Print("finished all layers without issues")
			return
		}
		playersLen, err := matchCrawler.Rdb.PlayersQueueLen()
		if err != nil {
			log.Print(err)
			log.Fatal("error getting len of player queue")
		}
		for playersLen > 0 {
			for i := 0; i < min(playersLen, matchCrawler.NumWorkers); i++ {
				currPuuid, err := matchCrawler.Rdb.DequeuePlayer()

				if err != nil {
					log.Print(err)
					log.Fatal("error poping from player queue")
				}
				log.Printf("getting matches for player: %s", currPuuid)
				matchCrawler.Wg.Add(1)
				go matchCrawler.GetMatchesFromPuuid(currPuuid)
			}
			matchCrawler.Wg.Wait()

			playersLen, err = matchCrawler.Rdb.PlayersQueueLen()
			if err != nil {
				log.Print(err)
				log.Fatal("error getting len of player queue")
			}
		}

		matchesLen, err := matchCrawler.Rdb.MatchesQueueLen()
		if err != nil {
			log.Print(err)
			log.Fatal("error getting len of match queue")
		}

		for matchesLen > 0 {
			for i := 0; i < min(matchesLen, matchCrawler.NumWorkers); i++ {
				matchId, err := matchCrawler.Rdb.DequeueMatch()
				if err != nil {
					log.Print(err)
					log.Fatal("error poping from matches queue")
				}
				log.Printf("getting match data for match: %s", matchId)
				matchCrawler.Wg.Add(1)
				go matchCrawler.GetMatchDataFromMatchID(matchId)
			}
			matchCrawler.Wg.Wait()

			idk, err := Pgdb.Client.BulkInsertUnits(Pgdb.Context, matchCrawler.CurrData)
			if err != nil {
				log.Fatal("error writing local data to database")
			}
			log.Printf("successfully inserted %d rows", idk)
			matchCrawler.CurrData = nil

			matchesLen, err = matchCrawler.Rdb.MatchesQueueLen()
			if err != nil {
				log.Print(err)
				log.Fatal("error getting len of matches queue")
			}
		}

		currDepth += 1
		log.Printf("finished %d layers without issues", currDepth)
	}

}
