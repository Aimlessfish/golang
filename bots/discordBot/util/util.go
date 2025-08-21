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

	err := godotenv.Load()
	if err != nil {
		logger.Info("INVALID BOT TOKEN", "ERROR", err.Error())
		return "broke mate"
	} else {
		token := os.Getenv("TOKEN")
		if len(token) == 0 {
			panic("Token length == 0!")
		} else {
			return token
		}
	}
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
