/*
API_CALLS for /sport/{SPORT}/
"soccer_epl" → English Premier League

"soccer_efl_championship" → EFL Championship

"soccer_fa_cup" → FA Cup

"soccer_england_league1" → League One

"soccer_england_league2" → League Two

*/

package api

import (
	loggerInit "discordBot/util"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type MatchOdds struct {
	HomeTeam     string
	AwayTeam     string
	CommenceTime string
	Bookmaker    string
	Odds         map[string]float64 // Outcome name -> price
}

type Outcome struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type Market struct {
	Key        string    `json:"key"`
	LastUpdate string    `json:"last_update"`
	Outcomes   []Outcome `json:"outcomes"`
}

type Bookmaker struct {
	Key        string   `json:"key"`
	Title      string   `json:"title"`
	LastUpdate string   `json:"last_update"`
	Markets    []Market `json:"markets"`
}

type Match struct {
	ID           string      `json:"id"`
	SportKey     string      `json:"sport_key"`
	SportTitle   string      `json:"sport_title"`
	CommenceTime string      `json:"commence_time"`
	HomeTeam     string      `json:"home_team"`
	AwayTeam     string      `json:"away_team"`
	Bookmakers   []Bookmaker `json:"bookmakers"`
}

func GetTheOddsAPI() ([]MatchOdds, error) {
	logger := loggerInit.LoggerInit("API", "THE_ODDS_API")
	err := godotenv.Load()
	if err != nil {
		logger.Info("INVALID BOT TOKEN", "ERROR", err.Error())
		return nil, err
	}
	token := os.Getenv("THE_ODDS")
	if len(token) == 0 {
		panic("Token length == 0!")
	}

	url := "https://api.the-odds-api.com/v4/sports/soccer_epl/odds?apiKey=" + token + "&regions=uk&markets=h2h"

	response, err := http.Get(url)
	if err != nil {
		logger.Error("Failed to get request", "error", err)
		os.Exit(1)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		logger.Error("Failed to read response body", "error", err)
		return nil, err
	}

	var matches []Match
	err = json.Unmarshal(body, &matches)
	if err != nil {
		logger.Error("Failed to parse JSON: ", "error", err)
		return nil, err
	}
	var results []MatchOdds

	for _, match := range matches {
		for _, bookmaker := range match.Bookmakers {
			for _, market := range bookmaker.Markets {
				if market.Key == "h2h" {
					odds := make(map[string]float64)
					for _, outcome := range market.Outcomes {
						odds[outcome.Name] = outcome.Price
					}
					results = append(results, MatchOdds{
						HomeTeam:     match.HomeTeam,
						AwayTeam:     match.AwayTeam,
						CommenceTime: match.CommenceTime,
						Bookmaker:    bookmaker.Title,
						Odds:         odds,
					})
				}
			}
		}
	}

	return results, nil
}
