package util

import (
	"log/slog"
	"os"
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
