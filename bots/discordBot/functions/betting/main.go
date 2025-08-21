package MatchOdds

import (
	the_odds "discordBot/functions/betting/api"
	"fmt"
	"log/slog"
	"os"

	"github.com/bwmarrin/discordgo"
)

func MatchOdds(server *discordgo.Session, message *discordgo.MessageCreate) error {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	response, err := the_odds.GetTheOddsAPI()
	if err != nil {
		return err
	}

	logger.Info("API Response Received")

	for _, match := range response {
		server.ChannelMessageSend(message.ChannelID, "--------------------------------------------------")
		server.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Match: %s vs %s\n", match.HomeTeam, match.AwayTeam))
		server.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Time: %s\n", match.CommenceTime))
		server.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Bookmaker: %s\n", match.Bookmaker))
		for outcome, price := range match.Odds {
			server.ChannelMessageSend(message.ChannelID, fmt.Sprintf("  %s: %.2f\n", outcome, price))
		}
	}
	return nil
}
