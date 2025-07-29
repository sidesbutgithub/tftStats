package crawler

import "log"

func (crawlerInst *Crawler) DequeueMatchEnqueuePlayers() {
	defer crawlerInst.Wg.Done()
	matchId, err := crawlerInst.Rdb.DequeueMatch()
	if err != nil {
		log.Print("error dequing matchId")
		log.Print(err)
		return
	}
	crawlerInst.GetMatchDataFromMatchID(matchId)
}
