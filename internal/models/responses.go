package models

type LeagueResponse struct {
	Tier     string `json:"tier"`
	LeagueID string `json:"leagueId"`
	Queue    string `json:"queue"`
	Name     string `json:"name"`
	Entries  []struct {
		Puuid        string `json:"puuid"`
		LeaguePoints int    `json:"leaguePoints"`
		Rank         string `json:"rank"`
		Wins         int    `json:"wins"`
		Losses       int    `json:"losses"`
		Veteran      bool   `json:"veteran"`
		Inactive     bool   `json:"inactive"`
		FreshBlood   bool   `json:"freshBlood"`
		HotStreak    bool   `json:"hotStreak"`
	} `json:"entries"`
}

type MatchResponse struct {
	Metadata struct {
		DataVersion  string   `json:"data_version"`
		MatchID      string   `json:"match_id"`
		Participants []string `json:"participants"`
	} `json:"metadata"`
	Info struct {
		EndOfGameResult string  `json:"endOfGameResult"`
		GameCreation    int64   `json:"gameCreation"`
		GameID          int64   `json:"gameId"`
		GameDatetime    int64   `json:"game_datetime"`
		GameLength      float64 `json:"game_length"`
		GameVersion     string  `json:"game_version"`
		MapID           int     `json:"mapId"`
		Participants    []struct {
			Companion struct {
				ContentID string `json:"content_ID"`
				ItemID    int    `json:"item_ID"`
				SkinID    int    `json:"skin_ID"`
				Species   string `json:"species"`
			} `json:"companion"`
			GoldLeft  int `json:"gold_left"`
			LastRound int `json:"last_round"`
			Level     int `json:"level"`
			Missions  struct {
				PlayerScore2 int `json:"PlayerScore2"`
			} `json:"missions"`
			Placement            int     `json:"placement"`
			PlayersEliminated    int     `json:"players_eliminated"`
			Puuid                string  `json:"puuid"`
			RiotIDGameName       string  `json:"riotIdGameName"`
			RiotIDTagline        string  `json:"riotIdTagline"`
			TimeEliminated       float64 `json:"time_eliminated"`
			TotalDamageToPlayers int     `json:"total_damage_to_players"`
			Traits               []struct {
				Name        string `json:"name"`
				NumUnits    int    `json:"num_units"`
				Style       int    `json:"style"`
				TierCurrent int    `json:"tier_current"`
				TierTotal   int    `json:"tier_total"`
			} `json:"traits"`
			Units []struct {
				CharacterID string   `json:"character_id"`
				ItemNames   []string `json:"itemNames"`
				Name        string   `json:"name"`
				Rarity      int      `json:"rarity"`
				Tier        int      `json:"tier"`
			} `json:"units"`
			Win bool `json:"win"`
		} `json:"participants"`
		QueueID        int    `json:"queueId"`
		QueueID0       int    `json:"queue_id"`
		TftGameType    string `json:"tft_game_type"`
		TftSetCoreName string `json:"tft_set_core_name"`
		TftSetNumber   int    `json:"tft_set_number"`
	} `json:"info"`
}

type ListMatches []string
