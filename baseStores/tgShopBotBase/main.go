package main

import (
	"log/slog"
	"os"
	api "telegramconnect/api"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	logger = slog.With("ID", "MAIN")

	response, err := api.ConnectAPI()
	if err != nil {
		logger.Error("Failed to connect.", "error", err.Error())
	} else {
		logger.Info("Connected to: " + response)
	}

}
