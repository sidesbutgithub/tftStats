package databaseClients

import (
	"context"
	"errors"
	"log"

	"github.com/redis/go-redis/v9"
)

type RedisDB struct {
	Client  *redis.Client
	Context context.Context
}

//Connect to Redis Database

func (db *RedisDB) ConnectRedis(redisHost string, redisPort string, redisPassword string, redisDBNum int) {
	db.Client = redis.NewClient(&redis.Options{
		Addr:     redisHost + ":" + redisPort,
		Password: redisPassword, // No password set
		DB:       redisDBNum,    // Use default DB
	})
	db.Context = context.Background()
	err := db.Client.Ping(db.Context).Err()
	if err != nil {
		log.Fatal("Failed to connect to Redis")
	}
}

//Set for visited Players

func (db *RedisDB) MarkPlayerVisited(Puuid string) error {
	if db.Client == nil {
		return errors.New("database not connected")
	}
	err := db.Client.SAdd(db.Context, "visitedPlayers", Puuid).Err()
	return err
}

func (db *RedisDB) CheckPlayerVisited(Puuid string) (bool, error) {
	if db.Client == nil {
		return false, errors.New("database not connected")
	}
	visited, err := db.Client.SIsMember(db.Context, "visitedPlayers", Puuid).Result()
	if err != nil {
		return false, err
	}
	return visited, err
}

//Set for visited matches

func (db *RedisDB) MarkMatchVisited(matchId string) error {
	if db.Client == nil {
		return errors.New("database not connected")
	}
	err := db.Client.SAdd(db.Context, "visitedMatches", matchId).Err()
	if err != nil {
		log.Print("Failed to write match to visited set")
		return err
	}
	return nil
}

func (db *RedisDB) CheckMatchVisited(matchId string) (bool, error) {
	if db.Client == nil {
		return false, errors.New("database not connected")
	}
	visited, err := db.Client.SIsMember(db.Context, "visitedMatches", matchId).Result()
	if err != nil {
		return false, err
	}
	return visited, err
}

// Queue for players to visit
func (db *RedisDB) EnqueuePlayer(Puuid string) error {
	if db.Client == nil {
		return errors.New("database not connected")
	}
	err := db.Client.LPush(db.Context, "playersQueue", Puuid).Err()
	return err
}

func (db *RedisDB) DequeuePlayer() (string, error) {
	if db.Client == nil {
		return "", errors.New("database not connected")
	}
	queueLen, err := db.Client.LLen(db.Context, "playersQueue").Result()
	if err != nil {
		return "", err
	}
	if queueLen == 0 {
		return "", errors.New("players queue is empty")
	}
	puuid, err := db.Client.RPop(db.Context, "playersQueue").Result()
	if err != nil {
		return "", err
	}
	return puuid, err
}

// Queue for matches to visit
func (db *RedisDB) EnqueueMatch(MatchId string) error {
	if db.Client == nil {
		return errors.New("database not connected")
	}
	err := db.Client.LPush(db.Context, "matchesQueue", MatchId).Err()
	return err
}

func (db *RedisDB) DequeueMatch() (string, error) {
	if db.Client == nil {
		return "", errors.New("database not connected")
	}
	queueLen, err := db.Client.LLen(db.Context, "matchesQueue").Result()
	if err != nil {
		return "", err
	}
	if queueLen == 0 {
		return "", errors.New("matches queue is empty")
	}
	puuid, err := db.Client.RPop(db.Context, "matchesQueue").Result()
	if err != nil {
		return "", err
	}
	return puuid, err
}

func (db *RedisDB) PlayersQueueLen() (int, error) {
	queueLen, err := db.Client.LLen(db.Context, "playersQueue").Result()
	if err != nil {
		log.Print("error in getting length of player queue")
		return 0, err
	}
	return int(queueLen), nil
}

func (db *RedisDB) MatchesQueueLen() (int, error) {
	queueLen, err := db.Client.LLen(db.Context, "matchesQueue").Result()
	if err != nil {
		log.Print("error in getting length of matches queue")
		return 0, err
	}
	return int(queueLen), nil
}
