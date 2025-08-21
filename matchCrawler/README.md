# TFT Match Crawler

Crawls matches of the strategy game Teamfight Tactics by Riot Games, storing statistics of game data provided by the official Riot API into a PostgreSQL database

## DequeuePlayerEnqueueMatches
Gets a match from the match queue, requests the match data from the Riot Games API (RGAPI), saves data about units into a PostgreSQL database and inserts the players from that match into the players queue if they have not been visited before

## DequeueMatchEnqueuePlayers
Gets a player from the players queue, requests their past 20 matches(if they are after the minTime) from the RGAPI and inserts the matches into the matches queue
