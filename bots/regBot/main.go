package main

import (
	"log/slog"
	"os"
	getproxy "regbot/proxyhandler"
	util "regbot/util"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	logger = slog.With("LogID", "REGBOT MAIN")

	userOS, err := util.CheckOS()
	if err != nil {
		logger.Error("Failed to check OS, Exiting.", "error", err)
		os.Exit(1)
	}
	logger.Info(userOS)

	driver, err := getproxy.GetProxy()
	if err != nil {
		logger.Error("Failed to init proxy.GetProxy", "internal error", err)
	}
	err = driver.Get("https://aimlessdev.co.uk")
	if err != nil {
		logger.Error("Failed to get url", "error", err)
		os.Exit(1)
	}

}
