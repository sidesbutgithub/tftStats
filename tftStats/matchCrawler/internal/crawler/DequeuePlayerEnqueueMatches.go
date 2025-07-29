package crawler

import "log"

func (crawlerInst *Crawler) DequeuePlayerEnqueueMatches() {
	defer crawlerInst.Wg.Done()
	currPuuid, err := crawlerInst.Rdb.DequeuePlayer()
	if err != nil {
		log.Print("error dequing playerId")
		log.Print(err)
		return
	}
	crawlerInst.GetMatchesFromPuuid(currPuuid)
}
