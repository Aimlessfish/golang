package apicalls

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
	WIN_GECKO_PATH   = "C:\\Users\\notWill\\Documents\\GitHub\\automation\\golang\\bots\\regBot\\bin\\geckodriver.exe"
	LINUX_GECKO_PATH = "../bin/geckodriver"
	GECKO_PORT       = 4444
	PROXY_SCRAPE_API = "https://api.proxyscrape.com/v4/free-proxy-list/get?request=displayproxies&protocol=http&timeout=10000&country=all&ssl=all&anonymity=all&skip=0&limit=500"
	PUB_PROXY_API    = "http://pubproxy.com/api/proxy?https=true&type=http&format=json"
)

type osType struct {
	Linux   string
	Windows string
	Mac     string
}

type PubProxyResponse struct {
	Data []struct {
		IPPort string `json:"ipPort"` // This is what we care about
		IP     string `json:"ip"`
		Port   string `json:"port"`
		// Other fields can be added if needed
	} `json:"data"`
	Count int `json:"count"`
}

type APIProvider struct {
	PubProxy    string
	ProxyScrape string
}

var OSLIST = []osType{
	{
		Linux:   "FireFox",
		Windows: "FireFox",
	},
	{
		Linux:   "FireFox",
		Windows: "FireFox",
	},
}

func loggerInit(logID, descriptor string) *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	logger = slog.With(logID, descriptor)
	return logger
}

func FetchFromPubProxy() ([]string, error) {
	logger := loggerInit("LogID", "FetchFromPubProxy")

	resp, err := http.Get(PUB_PROXY_API)
	if err != nil {
		logger.Error("Failed to build http request", "error", err)
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

func FetchProxyScrape() ([]string, error) {
	logger := loggerInit("LogID", "FetchProxyScrape")

	resp, err := http.Get(PROXY_SCRAPE_API)
	if err != nil {
		logger.Error("Failed to fetch proxies from ProxyScrape", "error", err)
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

	// Parse proxies line by line
	lines := strings.Split(strings.TrimSpace(string(body)), "\n")

	var proxies []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			proxies = append(proxies, line)
		}
	}

	if len(proxies) == 0 {
		err := fmt.Errorf("no proxies found in ProxyScrape response")
		logger.Error(err.Error())
		return nil, err
	}

	return proxies, nil
}
