package main

import (
	"log/slog"
	"net"
	"os"
	"time"

	"./apiCalls"
)

const (
	PROXY_SCRAPE_API = "https://api.proxyscrape.com/v4/free-proxy-list/get?request=displayproxies&protocol=http&timeout=10000&country=all&ssl=all&anonymity=all&skip=0&limit=500"
	PUB_PROXY_API    = "http://pubproxy.com/api/proxy?https=true&type=http&format=json"
)

// change this to getProxy )[]string, error) {}
// production only
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	logger = slog.With("logID", "copyFile")

	proxies, err := APICall()
	if err != nil {
		logger.Error("Failed to run APICall", "error", err)
	}

	workingProxies, err := TestProxy(proxies)
	if err != nil {
		logger.Error("Failed to test proxies", "error", err)
	}
	if len(workingProxies) == 0 {
		logger.Error("No working proxies found")
	}

	for _, proxy := range workingProxies {
		logger.Info(proxy)
	}

}

func APICall() ([]string, error) { // get proxies
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	logger = slog.With("logID", "proxyAPICALL")

	var allProxies []string

	pubProxies, err := apiCalls.FetchPubProxy(logger)
	if err == nil {
		allProxies = append(allProxies, pubProxies...)
		return allProxies, nil
	} else {
		logger.Error("Failed to fetch from PubProxy.com", "error", err)
	}
	if len(allProxies) != 0 {
		logger.Error("Failed to scrape PubProxy API")
		return nil, nil
	}
	proxies, err := apiCalls.FetchProxyScrape(logger)
	allProxies = append(allProxies, proxies...)
	if len(proxies) == 0 {
		logger.Error("Failed to get list from proxy scrape", "error", err)
	}
	if len(allProxies) == 0 {
		logger.Error("NO PROXIES FOUND FROM BOTH SITES CHECK FIREWALL / ISP services")
		os.Exit(1)
	}

	// Fallback: plain text)
	return allProxies, nil
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

		if len(workingProxies) == 1 { // EDIT THIS TO RETURN AS MANY AS YOU WANT
			logger.Info("Current Proxy: " + proxy)
			break
		}

	}

	return workingProxies, nil
}
