package main

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/sidesbutgithub/tftStats/matchCrawler/internal/crawler"
	"github.com/sidesbutgithub/tftStats/matchCrawler/internal/database"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Failed to load .env file")
	}

	//connect to postgres
	dbURI := os.Getenv("DB_URI")
	var Pgdb database.PostgresDB
	defer Pgdb.CloseConn()
	err = Pgdb.ConnectPostgres(dbURI)
	if err != nil {
		log.Fatal("Failed to connect to Postgres DB")
	}

	//connect to redis
	var Rdb database.RedisDB
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
		Queue:      &Rdb,
		DB:         &Pgdb,
		RiotApiKey: riotApiKey,
		NumWorkers: 10,
	}
	//first run of crawler on each container will be same puuid
	startingPuuid := os.Getenv("STARTING_PUUID")
	initialVisited, err := matchCrawler.Queue.CheckPlayerVisited(startingPuuid)
	if err != nil {
		log.Print(err)
		log.Fatal("error checking if initial player in visited")
	}

	if !initialVisited {
		matchCrawler.AddPlayer(os.Getenv("STARTING_PUUID"))
	}

	maxDepth := 3
	currDepth := 0
	//main loop
	for {

		if currDepth >= maxDepth {
			log.Print("finished all layers without issues")
			return
		}
		playersLen, err := matchCrawler.Queue.PlayersQueueLen()
		if err != nil {
			log.Print(err)
			log.Fatal("error getting len of player queue")
		}
		for playersLen > 0 {
			currPuuid, err := matchCrawler.Queue.DequeuePlayer()
			if err != nil {
				log.Print(err)
				log.Fatal("error poping from player queue")
			}
			err = matchCrawler.GetMatches(currPuuid)
			if err != nil {
				log.Print(err)
				log.Fatal("error getting matches for given playerID")
			}
			playersLen, err = matchCrawler.Queue.PlayersQueueLen()
			if err != nil {
				log.Print(err)
				log.Fatal("error getting len of player queue")
			}
		}
		matchesLen, err := matchCrawler.Queue.MatchesQueueLen()
		if err != nil {
			log.Print(err)
			log.Fatal("error getting len of match queue")
		}
		for matchesLen > 0 {
			currMatchId, err := matchCrawler.Queue.DequeueMatch()
			if err != nil {
				log.Print(err)
				log.Fatal("error poping from match queue")
			}
			err = matchCrawler.GetMatchData(currMatchId)
			if err != nil {
				log.Print(err)
				log.Fatal("error getting match data for given matchId")
			}
			matchesLen, err = matchCrawler.Queue.MatchesQueueLen()
			if err != nil {
				log.Print(err)
				log.Fatal("error getting len of match queue")
			}
		}
		currDepth += 1
		log.Printf("finished %d layers without issues", currDepth)
	}

}
