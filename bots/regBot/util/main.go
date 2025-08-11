package util

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"runtime"
)

func LoggerInit(logID, descriptor string) *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	logger = slog.With(logID, descriptor)
	return logger
}

func ServerInit(port string, logger *slog.Logger) (string, error) {
	logger = logger.With("ServerInit", "Utilities")
	userOS, err := checkOS()
	if err != nil {
		logger.Error("Failed to check OS, Exiting.", "error", err)
		os.Exit(1)
	}
	logger.Info(userOS)
	status, err := fireWall(port, logger)
	if err != nil || !status {
		logger.Error("Error firewal returned false", "error", err)
		return "", err
	}
	return userOS, nil
}

func checkOS() (string, error) {
	os := runtime.GOOS
	if os == "darwin" {
		msg := "FUCK OFF"
		panic(msg)
	}

	return os, nil
}

func fireWall(port string, logger *slog.Logger) (bool, error) {
	logger = logger.With("FireWall", "Utilities")

	cmd := exec.Command("ufw", "allow", port)

	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("Failed to run ufw allow", port, "error", err)
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Printf("Output: %s\n", output)

	return true, nil
}
