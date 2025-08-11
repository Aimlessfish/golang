package apiCalls

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	util "regbot/util"
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

func FetchPubProxy() ([]string, error) {
	logger := util.LoggerInit("LogID", "FetchPubProxy")
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

func FetchProxyScrape() ([]string, error) {
	logger := util.LoggerInit("LogID", "FetchProxyScrape")
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
	lines := strings.Split(strings.TrimSpace(string(body)), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			proxies = append(proxies, line)
		}
	}
	return proxies, nil
}
