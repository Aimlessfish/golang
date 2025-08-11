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
	"strings"
	"time"

	"apiCalls/apiCalls"

	"github.com/tebeka/selenium"
)

func GetProxiedSession(osType string) (selenium.WebDriver, *selenium.Service, error) {
	err := apiCalls.FetchFromPubProxy()
	logger := loggerInit("ID", "TestProxy")

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

	if osType != "linux" {
		if osType == "windows" && browserType == "FireFox" {
			service, err = selenium.NewGeckoDriverService(WIN_GECKO_PATH, GECKO_PORT)
			if err != nil {
				logger.Error("Failed to start geckodriver service", "error", err)
				return nil, nil, err
			}
		}

		driver, err = ApplyBrowserProxy(browserType, workingProxies[0])
		if err != nil {
			logger.Error("Failed to create proxied browser session", "error", err)
			if service != nil {
				service.Stop()
			}
			return nil, nil, err
		}

		return driver, service, nil
	}

	/* IF LINUX DEFAULT CASE */
	service, err = selenium.NewGeckoDriverService(LINUX_GECKO_PATH, GECKO_PORT)
	if err != nil {
		logger.Error("Failed to start geckodriver service", "error", err)
		return nil, nil, err
	}

	driver, err = ApplyBrowserProxy(browserType, workingProxies[0])
	if err != nil {
		logger.Error("Failed to create proxied browser session", "error", err)
		if service != nil {
			service.Stop()
		}
		return nil, nil, err
	}

	return driver, service, nil
}

func APICall(osType string) ([]string, error) { // get proxies
	logger := loggerInit("LogID", "APICall")

	var allProxies []string

	pubProxies, err := FetchFromPubProxy()
	if err == nil {
		allProxies = append(allProxies, pubProxies...)

	} else {
		logger.Error("Failed to fetch from PubProxy.com", "error", err)
	}

	// Fallback: plain text)
	return allProxies, nil
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
