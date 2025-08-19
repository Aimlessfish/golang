package main

import (
	"os"
	getproxy "regbot/proxyHandler"
	util "regbot/util"
)

const (
	GECKO_PORT = "5555"
)

func main() {
	logger := util.LoggerInit("MAIN", "MAIN")

	userOS, err := util.ServerInit(GECKO_PORT, logger)
	if err != nil {
		logger.Error("Server init failed", "error", err)
		os.Exit(1)
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
	err = driver.Get("http://ifconfig.me")

	// err = driver.Get("https://hpivaluations.com/#")
	if err != nil {
		logger.Error("Failed to get url", "error", err)
		os.Exit(1)
	}

}
