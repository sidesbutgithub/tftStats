package main

import (
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/sidesbutgithub/tftStats/matchCrawler/internal/crawler"
	"github.com/sidesbutgithub/tftStats/matchCrawler/internal/database"
	"github.com/sidesbutgithub/tftStats/matchCrawler/internal/databaseClients"
	"golang.org/x/time/rate"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Print("Failed to load .env file")
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
	RdbConnString := os.Getenv("REDIS_URI")

	Rdb.ConnectRedis(RdbConnString)
	if err != nil {
		log.Print(RdbConnString)
		log.Fatal("Failed to connect to Redis DB")
	}

	Rdb.SetTimeout(time.Minute)

	riotApiKey := os.Getenv("RIOT_API_KEY")

	mWorkers, err := strconv.Atoi(os.Getenv("MATCH_WORKERS"))
	if err != nil {
		log.Print("Error Parsing Number of Match Workers from env, defaulting to 5")
		mWorkers = 5
	}

	pWorkers, err := strconv.Atoi(os.Getenv("PLAYER_WORKERS"))
	if err != nil {
		log.Print("Error Parsing Number of Player Workers, defaulting to 2")
		pWorkers = 5
	}

	numRetries, err := strconv.Atoi(os.Getenv("MAX_RETRIES"))
	if err != nil {
		log.Print("Error Parsing Number of Retries allowed, defaulting to 5")
		numRetries = 5
	}

	matchCrawler := &crawler.Crawler{
		Mu:               &sync.Mutex{},
		Wg:               &sync.WaitGroup{},
		Rl:               rate.NewLimiter(rate.Limit(float64(100)/float64(120)), 1),
		Rdb:              &Rdb,
		CurrData:         make([]database.BulkInsertUnitsParams, 0),
		MatchesStartTime: os.Getenv("START_TIME"),
		RiotApiKey:       riotApiKey,

		MatchWorkers:  mWorkers,
		PlayerWorkers: pWorkers,
		MaxRetries:    numRetries,
	}
	//first run of crawler on each container will be same puuid
	//startingPuuid := os.Getenv("STARTING_PUUID")
	//matchCrawler.AddPlayerIfNotVisited(startingPuuid)

	//main loop
	for {
		for i := 0; i < matchCrawler.PlayerWorkers; i++ {
			matchCrawler.Wg.Add(1)
			go matchCrawler.DequeuePlayerEnqueueMatches()
		}
		for i := 0; i < matchCrawler.MatchWorkers; i++ {
			matchCrawler.Wg.Add(1)
			go matchCrawler.DequeueMatchEnqueuePlayers()
		}
		matchCrawler.Wg.Wait()
		rowsInserted, err := Pgdb.Client.BulkInsertUnits(Pgdb.Context, matchCrawler.CurrData)
		if err != nil {
			log.Fatal("error writing local data to database")
		}
		log.Printf("successfully inserted %d rows", rowsInserted)
		matchCrawler.CurrData = nil

		playersLen, err := matchCrawler.Rdb.PlayersQueueLen()
		if err != nil {
			log.Print("err getting players queue Length")
			log.Print(err)
		}
		log.Printf("Players left in queue: %d", playersLen)
		matchesLen, err := matchCrawler.Rdb.MatchesQueueLen()
		if err != nil {
			log.Print("err getting Matches queue Length")
			log.Print(err)
		}
		log.Printf("Matches left in queue: %d", matchesLen)

		numMatchesCrawled, err := matchCrawler.Rdb.Client.SCard(Rdb.Context, "visitedMatches").Result()
		if err != nil {
			log.Print("err getting number of visited Matches")
			log.Print(err)
		}

		log.Printf("Number of matches crawled: %d", int(numMatchesCrawled)-matchesLen)

	}

}
