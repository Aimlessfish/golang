package proxyHandler

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"regbot/proxyHandler/apiCalls"
	util "regbot/util"

	"github.com/tebeka/selenium"
)

const (
	WIN_GECKO_PATH   = "C:\\Users\\notWill\\Documents\\GitHub\\automation\\golang\\bots\\regBot\\bin\\geckodriver.exe"
	LINUX_GECKO_PATH = "./bin/geckodriver"
	GECKO_PORT       = 5555
	PROXY_SCRAPE_API = "https://api.proxyscrape.com/v4/free-proxy-list/get?request=displayproxies&protocol=http&timeout=10000&country=all&ssl=all&anonymity=all&skip=0&limit=500"
	PUB_PROXY_API    = "http://pubproxy.com/api/proxy?https=true&type=http&format=json"
)

var LINUX_FF_BINARY = filepath.Join("bin", "firefox", "firefox")

func GetProxiedSession(osType string) (selenium.WebDriver, *selenium.Service, error) {
	// proxies, err := apiCalls.FetchFromPubProxy()
	logger := util.LoggerInit("ID", "TestProxy")

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

	browserType, version, err := DetectBrowserAndVersion(osType)
	if err != nil {
		logger.Error("Failed to detect browser and version", "error", err)
		return nil, nil, err
	}
	logger.Info("Using browser", "type", browserType, "version", version)

	var service *selenium.Service
	var driver selenium.WebDriver
	proxy := workingProxies[0]

	if osType != "linux" {
		if osType == "windows" && browserType == "FireFox" {
			service, err = selenium.NewGeckoDriverService(WIN_GECKO_PATH, GECKO_PORT)
			if err != nil {
				logger.Error("Failed to start geckodriver service", "error", err)
				return nil, nil, err
			}
		}
		driver, err = BrowserProxyWindows(browserType, proxy)
		if err != nil {
			logger.Error("Failed to create proxied browser session", "error", err)
			if service != nil {
				service.Stop()
			}
			return nil, nil, err
		}
		gecko_port := strconv.Itoa(GECKO_PORT)
		driver, service, err := util.BrowserProxyLinux(LINUX_FF_BINARY, LINUX_GECKO_PATH, browserType, gecko_port, proxy, logger)
		if err != nil {
			logger.Error("Failed to provide browser session.", "error", err)
			panic(err)
		}

		return driver, service, nil
	}

	servicePort := strconv.Itoa(GECKO_PORT)
	driver, service, err = util.BrowserProxyLinux(LINUX_GECKO_PATH, LINUX_FF_BINARY, browserType, servicePort, proxy, logger)
	if err != nil {
		logger.Error("Failed to run Linux Browser", "error", err)
		os.Exit(1)
	}
	defer driver.Quit()
	defer service.Stop()
	return driver, service, nil
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

func DetectBrowserAndVersion(osType string) (string, string, error) {
	switch osType {
	case "windows":
		return "FireFox", "latest", nil
	case "linux":
		return "FireFox", "latest", nil
	case "mac":
		return "Chrome", "latest", nil
	default:
		return "", "", fmt.Errorf("unsupported OS: %s", osType)
	}
}

func BrowserProxyWindows(browserType, workingProxy string) (selenium.WebDriver, error) {
	logger := util.LoggerInit("ID", "ApplyBrowserProxy")

	profileDir, err := os.MkdirTemp("", "firefox-profile")
	if err != nil {
		logger.Error("Failed to make temp dir")
	}
	defer os.RemoveAll(profileDir)

	parts := strings.Split(workingProxy, ":")
	proxyHost := parts[0]
	proxyPort := parts[1]

	prefs := fmt.Sprintf(`
user_pref("general.useragent.override", "MyCustomUserAgent/1.0");
user_pref("network.proxy.type", 1);
user_pref("network.proxy.http", "%v");
user_pref("network.proxy.http_port", %v);
user_pref("network.proxy.ssl", "%v");
user_pref("network.proxy.ssl_port", %v);
`, proxyHost, proxyPort, proxyHost, proxyPort)

	err = os.WriteFile(filepath.Join(profileDir, "prefs.js"), []byte(prefs), 0644)
	if err != nil {
		log.Fatal(err)
	}

	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)
	filepath.Walk(profileDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		relPath, _ := filepath.Rel(profileDir, path)
		w, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(w, f)
		return err
	})
	zipWriter.Close()

	// Encode profile as base64
	encodedProfile := base64.StdEncoding.EncodeToString(buf.Bytes())

	// Set capabilities
	caps := selenium.Capabilities{
		"browserName":     "firefox",
		"firefox_profile": encodedProfile,
	}

	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%v", GECKO_PORT))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to geckodriver: %v", err)
	}

	return wd, nil
}
