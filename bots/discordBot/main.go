package main

import (
	bot "discordBot/bot"
	initlogger "discordBot/util"
)

func main() {
	logger := initlogger.LoggerInit("MAIN", "MAIN")

	err := bot.ConnectAPI(logger)
	if err != nil {
		panic(err)
	}

}
