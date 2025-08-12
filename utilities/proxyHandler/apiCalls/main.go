package apiCalls

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

const (
	PROXY_SCRAPE_API = "https://api.proxyscrape.com/v4/free-proxy-list/get?request=displayproxies&protocol=http&timeout=10000&country=all&ssl=all&anonymity=all&skip=0&limit=100"
	PUB_PROXY_API    = "http://pubproxy.com/api/proxy?https=true&type=http&format=json"
)

type PubProxyResponse struct {
	Data []struct {
		IPPort string `json:"ipPort"` // This is what we care about
		IP     string `json:"ip"`
		Port   string `json:"port"`
		// Other fields can be added if needed
	} `json:"data"`
	Count int `json:"count"`
}

func APICall() ([]string, error) { // get proxies
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	logger = slog.With("logID", "proxyAPICALL")

	var allProxies []string

	pubProxies, err := FetchPubProxy(logger)
	logger.Debug("Fetching from PubProxy.com")
	if err == nil {
		allProxies = append(allProxies, pubProxies...)
		return allProxies, nil
	} else {
		logger.Error("Failed to fetch from PubProxy.com", "error", err)
	}
	logger.Debug("Fetching from ProxyScrape.com")
	proxies, err := FetchProxyScrape(logger)
	if len(proxies) == 0 {
		logger.Error("Failed to get list from proxy scrape", "error", err)
		return nil, err
	}
	allProxies = append(allProxies, proxies...)
	if len(allProxies) == 0 {
		logger.Error("NO PROXIES FOUND FROM BOTH SITES CHECK FIREWALL / ISP services")
		return nil, err
	}
	return allProxies, nil
}

func FetchPubProxy(logger *slog.Logger) ([]string, error) {
	logger = logger.With("LogID", "FetchPubProxy")
	resp, err := http.Get(PUB_PROXY_API)
	if err != nil {
		logger.Error("Failed to build http request for PubProxy.com", "error", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("PubProxy response read failed: %w", err)
	}

	var jsonResp PubProxyResponse
	if err := json.Unmarshal(body, &jsonResp); err != nil {
		return nil, fmt.Errorf("PubProxy JSON unmarshal failed: %w", err)
	}

	var proxies []string
	for _, entry := range jsonResp.Data {
		if strings.Contains(entry.IPPort, ":") {
			proxies = append(proxies, entry.IPPort)
		}
	}

	return proxies, nil
}

func FetchProxyScrape(logger *slog.Logger) ([]string, error) {
	logger = logger.With("LogID", "FetchProxyScrape")
	var proxies []string

	resp, err := http.Get(PROXY_SCRAPE_API)
	if err != nil {
		logger.Error("Failed to buld http request for ProxyScrape", "error", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("Unexpected HTTP status from ProxyScrape", "status_code", resp.StatusCode)
		return nil, fmt.Errorf("unexpected HTTP status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read response body", "error", err)
		return nil, fmt.Errorf("ProxyScrape response read failed: %w", err)
	}
	rawString := string(body)
	lines := strings.SplitSeq(rawString, "\n")
	for line := range lines {
		if line != "" {
			trimmed := strings.Trim(line, "\r")
			proxies = append(proxies, trimmed)
		}
	}

	return proxies, nil
}
