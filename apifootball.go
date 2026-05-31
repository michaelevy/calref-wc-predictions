package main

func eeeeeee() {}

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"net/http"
// 	"time"
// )

// const apiFootballBase = "https://v3.football.api-sports.io"

// type apiFixturesResp struct {
// 	Response []struct {
// 		Fixture struct {
// 			ID     int       `json:"id"`
// 			Date   time.Time `json:"date"`
// 			Status struct {
// 				Short string `json:"short"`
// 			} `json:"status"`
// 		} `json:"fixture"`
// 		Teams struct {
// 			Home struct {
// 				ID int `json:"id"`
// 			} `json:"home"`
// 			Away struct {
// 				ID int `json:"id"`
// 			} `json:"away"`
// 		} `json:"teams"`
// 		Goals struct {
// 			Home *int `json:"home"`
// 			Away *int `json:"away"`
// 		} `json:"goals"`
// 	} `json:"response"`
// }

// func fetchFixtures(ctx context.Context, apiKey string, league, season int) (*apiFixturesResp, error) {
// 	url := fmt.Sprintf("%s/fixtures?league=%d&season=%d", apiFootballBase, league, season)
// 	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
// 	if err != nil {
// 		return nil, err
// 	}
// 	req.Header.Set("x-apisports-key", apiKey)
// 	resp, err := http.DefaultClient.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()
// 	if resp.StatusCode != http.StatusOK {
// 		return nil, fmt.Errorf("api-football HTTP %d", resp.StatusCode)
// 	}
// 	var out apiFixturesResp
// 	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
// 		return nil, err
// 	}
// 	return &out, nil
// }
