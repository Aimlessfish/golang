package util

import (
	"log/slog"
	"os"
	"os/exec"
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

// SplitArgs splits a command string into arguments, handling quoted strings.
func SplitArgs(input string) []string {
	var args []string
	var current string
	inQuotes := false
	for i := 0; i < len(input); i++ {
		c := input[i]
		switch c {
		case ' ':
			if inQuotes {
				current += string(c)
			} else if len(current) > 0 {
				args = append(args, current)
				current = ""
			}
		case '"':
			inQuotes = !inQuotes
		default:
			current += string(c)
		}
	}
	if len(current) > 0 {
		args = append(args, current)
	}
	return args
}
func ExecReportBinary(command string, args ...string) (string, error) {
	logger := LoggerInit("UTIL", "ExecReportBinary")
	cmd := "./reporter/csreport"
	binaryCmd := exec.Command(cmd, append([]string{command}, args...)...)
	output, err := binaryCmd.CombinedOutput()
	if err != nil {
		logger.Error("Failed to execute report binary", "error", err, "output", string(output))
		return string(output), err
	}
	return string(output), nil
}

// func ExecReportBinary(command string, args ...string) error {
// 	logger := LoggerInit("UTIL", "ExecReportBinary")
// 	cmd := "./reporter/csreport"
// 	binaryCmd := exec.Command(cmd, append([]string{command}, args...)...)
// 	output, err := binaryCmd.CombinedOutput()
// 	if err != nil {
// 		logger.Error("Failed to execute report binary", "error", err, "output", string(output))
// 		return err
// 	}
// 	return nil
// }
