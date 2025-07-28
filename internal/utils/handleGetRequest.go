package utils

import (
	"errors"
	"io"
	"log"
	"net/http"
	"time"
)

// Returns the byte string
func HandleHttpGetReqWithRetries(reqAddress string, maxReqRetries int, maxBodyRetries int) ([]byte, error) {
	var res *http.Response
	var err error
	currRetries := 0
	for {
		res, err = http.Get(reqAddress)
		if err != nil {
			log.Print("Failed to get match response")
			log.Print(reqAddress)
			return nil, err
		}
		if res.StatusCode == 200 {
			defer res.Body.Close()
			b, err := io.ReadAll(res.Body)
			bodyRetries := 0
			for err != nil {
				bodyRetries += 1
				if bodyRetries > maxBodyRetries {
					log.Print(reqAddress)
					log.Printf("body read failed excessively")
					return nil, errors.New("max body read retries exceeded")
				}
				log.Print("error reading body data, retrying")
				b, err = io.ReadAll(res.Body)
			}
			return b, nil
		}

		if res.StatusCode == 429 {
			log.Print("program rate limiting failed")
			currRetries += 1
			if currRetries == 1 {
				log.Print("hit lower rate limit 20 reqs/s, sleeping 1s before retrying")
				time.Sleep(time.Second)
				log.Print("awake, retrying http request")
				continue
			} else if currRetries < maxReqRetries {
				log.Print(reqAddress)
				log.Print("hit greater rate limit 100 reqs/2 mins, sleeping 2 mins before retrying")
				time.Sleep(2 * time.Minute)
				log.Print("awake, retrying http request")
				continue
			} else {
				log.Print("hit rate limit max number of times, no more retying, quiting program")
				return nil, errors.New("rate limit exceeded excessively")
			}
		}
		log.Print(reqAddress)
		log.Print(res.StatusCode)
		log.Print("unexpected http status code")
		return nil, nil
	}
}
