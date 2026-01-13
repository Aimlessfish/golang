package main

import (
	"log/slog"
	"os"

	api "telegramconnect/api"

	"github.com/joho/godotenv"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	logger = slog.With("ID", "MAIN")

	// Load environment variables once at startup
	err := godotenv.Load()
	if err != nil {
		logger.Error("Failed to load .env file", "error", err)
		os.Exit(1)
	}

	response, err := api.ConnectAPI()
	if err != nil {
		logger.Error("Failed to connect.", response, err.Error())
		os.Exit(1)
	}
}
