package getproxy

import (
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/tebeka/selenium"
)

type BrowserType struct {
	FireFox string
	Chrome  string
}

func loggerInit(logID, descriptor string) *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	logger = slog.With(logID, descriptor)
	return logger
}

func GetProxy() (selenium.WebDriver, error) {
	logger := loggerInit("ID", "TestProxy")

	proxies, err := APICall()
	if err != nil {
		logger.Error("Failed to run APICall", "error", err)
	}

	workingProxies, err := TestProxy(proxies)
	if err != nil {
		logger.Error("Failed to test proxies", "error", err)
	} else if len(workingProxies) == 0 {
		logger.Error("working proxies returned 0", "panicking", err)
		os.Exit(1)
	}

	for _, proxy := range workingProxies {
		logger.Info("your proxy:" + proxy)
	}

	rawProxies, err := APICall()
	if err != nil {
		logger.Error("Failed to call API call from main", "error", err)
		os.Exit(1)
	}

	testedProxy, err := TestProxy(rawProxies)
	if err != nil {
		logger.Error("Failed testing proxies", "error", err)
		os.Exit(1)
	}

	driver, err := ApplyBrowserProxy("FireFox", string(testedProxy[0]))
	if err != nil {
		logger.Error("Failed to apply proxy to browser", "error", err)
		os.Exit(1)
	}

	return driver, nil

}

func APICall() ([]string, error) {
	logger := loggerInit("LogID", "APICall")

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.proxyscrape.com/v4/free-proxy-list/get?request=displayproxies&protocol=http&timeout=10000&country=all&ssl=all&anonymity=all&skip=0&limit=500", nil)
	if err != nil {
		logger.Error("Failed to create HTTP Request: connection issue?", "error", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Failed to action the request", "error", err)
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to parse response", "error", err)
	}
	proxies := strings.Split(string(bodyText), "\n")
	return proxies, nil
}

func TestProxy(proxies []string) ([]string, error) {
	var workingProxies []string
	logger := loggerInit("ID", "TestProxy")

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
			break
		}

	}

	return workingProxies, nil
}

func ApplyBrowserProxy(browserType, workingProxy string) (selenium.WebDriver, error) {
	logger := loggerInit("ID", "Apply Single Proxy")
	if len(workingProxy) == 0 {
		logger.Error("The passed working proxy value is nil", "exiting", workingProxy)
		os.Exit(1)
	}
	if len(browserType) == 0 {
		logger.Error("The passed browser value is nil", "exiting", browserType)
		os.Exit(1)
	}

	seleniumServer := "http://localhost:4444/wd/hub"

	var caps selenium.Capabilities

	if browserType == "FireFox" {
		caps = selenium.Capabilities{
			"browserName": "firefox",
		}
	}
	firefoxCaps := map[string]interface{}{
		"proxy": map[string]interface{}{
			"proxyType": "manual",
			"httpProxy": workingProxy,
			"sslProxy":  workingProxy,
		},
	}
	caps["moz:firefoxOptions"] = firefoxCaps

	// if browserType == "Chrome" {

	// }

	driver, err := selenium.NewRemote(caps, seleniumServer)
	if err != nil {
		logger.Error("Failed to create firefox driver", "error", err)
		os.Exit(1)
	}

	return driver, nil
}
