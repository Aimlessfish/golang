package proxyHandler

import (
	"fmt"
	"net"
	"os"
	"time"

	"regbot/proxyHandler/apiCalls"
	util "regbot/util"

	"github.com/tebeka/selenium"
)

const (
	WIN_GECKO_PATH   = "C:\\Users\\notWill\\Documents\\GitHub\\automation\\golang\\bots\\regBot\\bin\\geckodriver.exe"
	LINUX_GECKO_PATH = "./bin/geckodriver"
	LINUX_FF_BINARY  = "./bin/firefox/firefox"
	GECKO_PORT       = 5555
	PORT_STRING      = "5555"
	PROXY_SCRAPE_API = "https://api.proxyscrape.com/v4/free-proxy-list/get?request=displayproxies&protocol=http&timeout=10000&country=all&ssl=all&anonymity=all&skip=0&limit=500"
	PUB_PROXY_API    = "http://pubproxy.com/api/proxy?https=true&type=http&format=json"
)

func GetProxiedSession(osType string) (selenium.WebDriver, *selenium.Service, error) {
	// proxies, err := apiCalls.FetchFromPubProxy()
	logger := util.LoggerInit("ID", "Get Proxied Session")

	proxies, err := APICall(osType)
	if err != nil {
		logger.Error("Failed to run APICall", "error", err)
		return nil, nil, err
	}

	workingProxies, err := TestProxy(proxies)
	if err != nil {
		logger.Error("Failed to test proxies", "error", err)
		return nil, nil, err
	}
	if len(workingProxies) == 0 {
		logger.Error("No working proxies found")
		return nil, nil, fmt.Errorf("no working proxies found")
	}
	var service *selenium.Service
	var driver selenium.WebDriver
	proxy := workingProxies[0]

	if osType != "linux" {
		service, err = selenium.NewGeckoDriverService(WIN_GECKO_PATH, GECKO_PORT)
		if err != nil {
			logger.Error("Failed to start geckodriver service", "error", err)
			return nil, service, err
		}
		driver, err = util.BrowserProxyWindows(PORT_STRING, proxy, logger)
		if err != nil {
			logger.Error("Failed to create proxied browser session", "error", err)
			if service != nil {
				service.Stop()
			}
			return driver, service, err
		}
	}

	driver, err = util.BrowserProxyLinux(LINUX_FF_BINARY, LINUX_GECKO_PATH, PORT_STRING, proxy, logger)
	if err != nil {
		logger.Error("Failed to run Linux Browser", "error", err)
		os.Exit(1)
	}
	defer driver.Quit()
	defer service.Stop()
	return driver, nil, nil
}

func APICall(osType string) ([]string, error) { // get proxies
	logger := util.LoggerInit("LogID", "APICall")

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
	logger := util.LoggerInit("ID", "TestProxy")

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
