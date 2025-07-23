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

	matchCrawler.AddPlayer(os.Getenv("STARTING_PUUID"))

	//sample player
	currPuuid, err := matchCrawler.Queue.DequeuePlayer()
	if err != nil {
		log.Print(err)
		log.Fatal("failed to dequeue playerID")
	}
	log.Print(currPuuid)
	matchCrawler.GetMatches(currPuuid)
	for i := 0; i < 20; i++ {
		currMatch, err := matchCrawler.Queue.DequeueMatch()
		if err != nil {
			log.Print(err)
			log.Print(i)
			log.Fatal("failed to dequeue matchID")
		}
		err = matchCrawler.GetMatchData(currMatch)
		if err != nil {
			log.Print(err)
			log.Print(i)
			log.Print(currMatch)
			log.Fatal("failed to get match data for matchID")
		}
	}

	log.Print("finished without issues")
	//main loop
}
