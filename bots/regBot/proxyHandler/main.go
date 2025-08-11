package proxyhandler

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tebeka/selenium"
)

const (
	GECKO_PATH = "C:\\Users\\notWill\\Documents\\GitHub\\automation\\golang\\bots\\regBot\\bin\\geckodriver.exe"
	GECKO_PORT = 4444
)

var service *selenium.Service

type osType struct {
	Linux   string
	Windows string
	Mac     string
}

var OSLIST = []osType{
	{
		Linux:   "FireFox",
		Windows: "Chrome",
		Mac:     "Safari",
	},
	{
		Linux:   "Chrome",
		Windows: "FireFox",
		Mac:     "Chrome",
	},
}

func loggerInit(logID, descriptor string) *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	logger = slog.With(logID, descriptor)
	return logger
}

func GetProxiedSession(osType string) (selenium.WebDriver, *selenium.Service, error) {
	logger := loggerInit("ID", "TestProxy")

	proxies, err := APICall()
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

	if osType != "linux" && osType != "mac" {
		osType = "windows"
	}

	var service *selenium.Service
	if osType == "windows" && browserType == "FireFox" {
		service, err = selenium.NewGeckoDriverService(GECKO_PATH, GECKO_PORT)
		if err != nil {
			logger.Error("Failed to start geckodriver service", "error", err)
			return nil, nil, err
		}
	}

	driver, err := ApplyBrowserProxy(browserType, workingProxies[0])
	if err != nil {
		logger.Error("Failed to create proxied browser session", "error", err)
		if service != nil {
			service.Stop()
		}
		return nil, nil, err
	}

	return driver, service, nil
}

func APICall() ([]string, error) { // get proxies
	logger := loggerInit("LogID", "APICall")

	client := &http.Client{Timeout: 10 * time.Second} /* NEED A PAID API ALL THESE ARE SHIT*/
	req, err := http.NewRequest("GET", "https://api.proxyscrape.com/v4/free-proxy-list/get?request=displayproxies&protocol=http&timeout=10000&country=all&ssl=all&anonymity=all&skip=0&limit=500", nil)
	if err != nil {
		logger.Error("Failed to create HTTP Request", "error", err)
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Failed to perform HTTP request", "error", err)
		return nil, err
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read response", "error", err)
		return nil, err
	}
	proxies := strings.Split(strings.TrimSpace(string(bodyText)), "\n")
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

func ApplyBrowserProxy(browserType, workingProxy string) (selenium.WebDriver, error) {
	logger := loggerInit("ID", "ApplyBrowserProxy")

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

	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d", GECKO_PORT))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to geckodriver: %w", err)
	}

	return wd, nil
}
