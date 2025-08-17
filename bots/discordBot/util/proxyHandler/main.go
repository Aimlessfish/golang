/* Written by Will Meekin */
/* needs modifying for this project */
/* This scrapes two proxy api providers and returns as many working proxies as possile */

package proxyHandler

import (
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"

	"proxyHandler/apiCalls"
)

func ProxyHandler() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	logger = slog.With("logID", "Main")

	mode := flag.Int("mode", 0, "return type 0 == PROXY LIST || 1 == 1 PROXY")
	flag.Parse()
	if *mode > 1 {
		logger.Error("Options are 0 for list of proxy || 1 single proxy")
		os.Exit(1)
	}

	proxies, err := apiCalls.APICall(*mode)
	if err != nil {
		logger.Error("Failed to run APICall", "error", err)
	}
	workingProxies := testAndList(proxies)
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

func testAndList(proxies []string) []string {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	logger = slog.With("logID", "Test and List")

	workingProxies, err := TestProxy(proxies)
	if err != nil {
		logger.Error("Failed to test proxies", "error", err)
	}
	if len(workingProxies) == 0 {
		logger.Error("No working proxies found", "error", err)
	}
	return workingProxies
}
