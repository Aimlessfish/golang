package main

import (
	"log/slog"
	"os"
	getproxy "regbot/proxyHandler"
	util "regbot/util"
	"time"
)

const (
	GECKO_PORT = "4444"
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

	if userOS == "linux" {
		util.FireWall(GECKO_PORT, logger)
	}

	driver, service, err := getproxy.GetProxiedSession(userOS)
	if err != nil {
		logger.Error("Failed to init proxy.GetProxy", "internal error", err)
		os.Exit(1)
	}

	defer func() {
		if driver != nil {
			driver.Quit()
			logger.Info("Driver has quit")
		}
		if service != nil {
			service.Stop()
			logger.Info("Service has quit")
		}
	}()
	err = driver.Get("https://hpivaluations.com/#")
	if err != nil {
		logger.Error("Failed to get url", "error", err)
		os.Exit(1)
	}
	time.Sleep(20 * time.Second)
	driver.Quit()

}
