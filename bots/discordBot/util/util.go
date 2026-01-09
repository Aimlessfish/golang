package util

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"

	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

const (
	STEAM_MARKET_PRICE_OVERVIEW_URL = "https://steamcommunity.com/market/priceoverview/?appid=730&currency=2&market_hash_name="
)

func GetToken() string {
	logger := LoggerInit("GET BOT", "BOT")
	logger.Info("Getting bot token")

	// Load .env from project root only
	err := godotenv.Load()
	if err != nil {
		logger.Info("INVALID BOT TOKEN", "ERROR", err.Error())
		return "broke mate"
	} else {
		token := os.Getenv("DISCORD_BOT_TOKEN")
		if len(token) == 0 {
			panic("Token length == 0!")
		} else {
			return token
		}
	}
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
func ExecBinary(binaryPath string, command string, args ...string) (string, error) {
	logger := LoggerInit("UTIL", "ExecBinary")
	binaryCmd := exec.Command(binaryPath, append([]string{command}, args...)...)
	output, err := binaryCmd.CombinedOutput()
	if err != nil {
		logger.Error("Failed to execute binary", "error", err, "output", string(output))
		return string(output), err
	}
	return string(output), nil
}

// execCommandOutput runs a shell command and returns its output as a string
func FormatForSteamMarketInjection(input string) (string, error) {
	formatted := fmt.Sprintf(STEAM_MARKET_PRICE_OVERVIEW_URL+"%s", input)
	if formatted == "" {
		return "", fmt.Errorf("failed to format Steam Market Injection URL")
	}
	return formatted, nil
}

// execCommandOutput runs a shell command and returns its output as a string

func ExecCommandOutput(cmd string) (string, error) {
	out, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func MarketHashName(param string) string {
	return url.QueryEscape(param)
}

func ParseJsonOutput(output string) (map[string]interface{}, error) {
	logger := LoggerInit("UTIL", "ParseJsonOutput")
	var result map[string]interface{}
	err := json.Unmarshal([]byte(output), &result)
	if err != nil {
		logger.Error("Failed to parse JSON output", "error", err)
		return nil, err
	}
	return result, nil
}
