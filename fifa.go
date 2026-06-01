package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const fifaMatchesURL = "https://api.fifa.com/api/v3/calendar/matches?idCompetition=17&idSeason=285023&count=200&language=en"

type fifaMatchesResp struct {
	Results []struct {
		IdMatch              string    `json:"IdMatch"`
		Date                 time.Time `json:"Date"`
		MatchStatus          int       `json:"MatchStatus"`
		Home                 fifaSide  `json:"Home"`
		Away                 fifaSide  `json:"Away"`
		HomeTeamScore        *int      `json:"HomeTeamScore"`
		AwayTeamScore        *int      `json:"AwayTeamScore"`
		HomeTeamPenaltyScore *int      `json:"HomeTeamPenaltyScore"`
		AwayTeamPenaltyScore *int      `json:"AwayTeamPenaltyScore"`
	} `json:"Results"`
}

type fifaSide struct {
	IdCountry string `json:"IdCountry"`
	Score     *int   `json:"Score"`
}

func (a *App) fetchFixtures(ctx context.Context) (*fifaMatchesResp, error) {
	url := fmt.Sprintf(fifaMatchesURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	req.Header.Set("User-Agent", a.cfg.FIFAUserAgent)

	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fifa HTTP %d", resp.StatusCode)
	}
	var out fifaMatchesResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}
