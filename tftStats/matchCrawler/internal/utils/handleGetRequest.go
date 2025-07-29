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
	currRetries := 1
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
			if currRetries < maxReqRetries {
				log.Printf("number of consecutive retries: %d", currRetries)
				log.Printf("missmatch of program rate limit and riot rate limit, sleeping %ds before retrying", currRetries)
				time.Sleep(time.Second * time.Duration(currRetries))
				log.Print("awake, retrying http request")
				currRetries *= 2
				continue
			} else {
				log.Print(reqAddress)
				log.Print("hit rate limit max number of times, no more retying, quiting program")
				return nil, errors.New("rate limit exceeded excessively")
			}
		}
		//dont actually log address in prod cuz it contains api key
		log.Print(reqAddress)
		log.Print(res.StatusCode)
		log.Print("unexpected http status code, skipping current address")
		return nil, nil
	}
}
