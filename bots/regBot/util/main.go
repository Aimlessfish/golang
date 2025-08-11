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

func CheckOS() (string, error) {
	os := runtime.GOOS
	if os == "darwin" {
		msg := "FUCK OFF"
		panic(msg)
	}

	return os, nil
}

func BrowserInit() error {

	return nil
}
func FireWall(port string, logger *slog.Logger) (bool, error) {
	logger = logger.With("FireWall:", "utilities")

	cmd := exec.Command("ufw", "allow", port)

	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("Failed to run ufw allow", port, "error", err)
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Printf("Output: %s\n", output)

	return true, nil
}
