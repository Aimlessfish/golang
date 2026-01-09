/* Written by Will Meekin */
/* needs modifying for this project */
/* This scrapes two proxy api providers and returns as many working proxies as possile */

package proxy

import (
	"discordBot/util"
	"encoding/json"
)

const (
	binaryPath = "./bin/proxy"
)

type OutputData struct {
	Proxies []string `json:"proxies"`
}

type ProxyEntry struct {
	IP   string `json:"ip"`
	Port string `json:"port"`
}

func ProxyHandler(proxyType string) []string {
	logger := util.LoggerInit("PROXY HANDLER", "PROXY")

	mode := proxyType
	var output string
	var err error

	switch mode {
	case "http":
		output, err = util.ExecBinary(binaryPath, "http")
	case "https":
		output, err = util.ExecBinary(binaryPath, "https")
	case "socks5":
		output, err = util.ExecBinary(binaryPath, "socks5")
	default:
		logger.Error("Invalid proxy type", "type", proxyType)
		return []string{}
	}

	if err != nil {
		logger.Error("Failed to execute proxy binary", "error", err, "type", proxyType)
		return []string{}
	}

	var proxies []ProxyEntry
	err = json.Unmarshal([]byte(output), &proxies)
	if err != nil {
		logger.Error("Failed to parse JSON output", "error", err, "type", proxyType)
		return []string{}
	}

	var result []string
	for _, proxy := range proxies {
		result = append(result, proxy.IP+":"+proxy.Port)
	}

	return result
}
