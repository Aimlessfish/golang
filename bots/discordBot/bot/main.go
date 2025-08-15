package bot

import (
	initlogger "golang/utilities/initLogger"
	"os"
)

func GetBotToken() (string, error) {
	logger := initlogger.LoggerInit("GET BOT", "BOT")
	logger.Info("Getting bot token")
	token := os.Getenv("BOT_TOKEN")
	if token != "" {
		logger.Info("INVALID BOT TOKEN", "ERROR", token)
		return "INVALID BOT TOKEN", nil

	}
	return token, nil
}
