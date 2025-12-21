package util

import (
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

func GetToken() string {
	logger := LoggerInit("GET BOT", "BOT")
	logger.Info("Getting bot token")

	// Load .env from project root only
	err := godotenv.Load()
	if err != nil {
		logger.Info("No .env file found", "ERROR", err.Error())
	}

	// Try DISCORD_BOT_TOKEN first, then TOKEN (legacy format)
	token := os.Getenv("DISCORD_BOT_TOKEN")
	if len(token) == 0 {
		token = os.Getenv("TOKEN")
	}

	if len(token) == 0 {
		panic("Token not found! Please set DISCORD_BOT_TOKEN or TOKEN in .env file")
	}

	return token
}

func GetEnvAsInt(key string, defaultVal int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultVal
}

func GetEnvAsBool(key string, defaultVal bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultVal
}

func LoadEnv() error {

	// Load .env from project root only
	return godotenv.Load()
}

func LoggerInit(logID, descriptor string) *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	logger = slog.With(logID, descriptor)
	return logger
}

func MessageTTL(msgID string) (bool, error) {
	logger := LoggerInit("UTIL", "MessageTLL")
	const discordEpoch = 1420070400000

	id64, err := strconv.ParseInt(msgID, 10, 64)
	if err != nil {
		logger.Error("Failed to parse Message Date from msg.ID", "error", err)
		os.Exit(1)
	}

	timestamp := (id64 >> 22) + discordEpoch
	messageTime := time.UnixMilli(timestamp)

	if time.Since(messageTime) > (14 * 24 * time.Hour) {
		return false, nil
	}

	return true, nil
}
