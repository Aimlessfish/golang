/* Written by Will Meekin */
/* This scrapes two proxy api providers and returns as many working proxies as possile */
package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"

	"proxyHandler/apiCalls"
)

// production only

/*
type flags struct {
	returnType int
}
*/

/* change this to getProxy() ([]string, error) {
	flag.int("type", 0, "return type 0 == 1 PROXY || 1 == LIST OF PROXIES")
	*** main logic here ****

	return proxies, nil
}*/

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	logger = slog.With("logID", "Main")

	proxies, err := apiCalls.APICall()
	if err != nil {
		logger.Error("Failed to run APICall", "error", err)
	}

	fmt.Println("Testing....")
	workingProxies, err := TestProxy(proxies)
	if err != nil {
		logger.Error("Failed to test proxies", "error", err)
	}
	if len(workingProxies) == 0 {
		logger.Error("No working proxies found", "error", err)
	}
	fmt.Println("Output:")
	for _, proxy := range workingProxies {
		fmt.Println(proxy)
	}

}

func TestProxy(proxies []string) ([]string, error) {
	var workingProxies []string
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	logger = slog.With("logID", "TestProxy")

	for _, proxy := range proxies {
		timeout := 2 * time.Second
		conn, err := net.DialTimeout("tcp", "google.com:80", timeout)
		if err != nil {
			logger.Error("Ping failed for "+proxy, "error", err)
			continue
		}
		conn.Close()

		workingProxies = append(workingProxies, proxy)

	}

	return workingProxies, nil
}
