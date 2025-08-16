package main

import (
	bot "golang/bots/discordBot/bot"
	initlogger "golang/utilities/initLogger"
)

func main() {
	logger := initlogger.LoggerInit("MAIN", "MAIN")

	err := bot.ConnectAPI(logger)
	if err != nil {
		panic(err)
	}

}
