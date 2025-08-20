package main

import (
	the_odds "discordBot/functions/betting/api"
	"fmt"
	"log/slog"
	"os"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	response, err := the_odds.GetTheOddsAPI()
	if err != nil {
		panic(err)
	}

	logger.Info("API Response Received")

	for _, match := range response {
		fmt.Println("--------------------------------------------------")
		fmt.Printf("Match: %s vs %s\n", match.HomeTeam, match.AwayTeam)
		fmt.Printf("Time: %s\n", match.CommenceTime)
		fmt.Printf("Bookmaker: %s\n", match.Bookmaker)
		for outcome, price := range match.Odds {
			fmt.Printf("  %s: %.2f\n", outcome, price)
		}
	}
}
