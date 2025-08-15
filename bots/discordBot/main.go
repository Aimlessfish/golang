package discordbot

import (
	initlogger "golang/utilities/initLogger"
)

func main() {
	logger := initlogger.LoggerInit("MAIN", "MAIN")

	logger.Debug("DiscordBot! Started")

}
