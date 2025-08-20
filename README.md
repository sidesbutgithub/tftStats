# Teamfight Tactics(TFT) Stats
An application to gather statistics about units from the Riot Games API for the strategy game Teamfight Tactics


## Prerequisites
1. A Riot Games API Key
2. git for cloning the repo
3. Docker for running the project
While it is technically possible to run each of the components individually and connect them, it is much more tedious than using docker and will not be documented

## Set-up

### Installation
Clone the repo with `git clone https://github.com/sidesbutgithub/tftStats.git` and cd into it with `cd tftStats` before running the whole application using docker-compose with you desired flags
```
git clone https://github.com/sidesbutgithub/tftStats.git
cd tftStats
docker-compose up --build --scale matchcrawler=<your desired number of crawler containers> <-d if you want it to run detatched>
```
### .env
create a file named `.env` in the root directory of the project with the following environment variables to be used by the program
```
RIOT_API_KEY = <your Riot Games API Key>
DB_URI = 
REDIS_URI
MAX_RETRIES = <the max amount of retries for various processes like riot api requests or body reads>
START_TIME = <the minimum unix time stamp of the matches you want included, for example the time stamp of the last patch release>
MATCH_WORKERS = <number of goroutines processing matches for each matchcrawler container
PLAYER_WORKERS = <number of goroutines processing players for each matchcrawler conatiner>
```
### Running
use docker-compose to run all the services
```
docker-compose up --build --scale matchcrawler=<desired number of match crawler containers> -d
```

## Routes

### TODO
