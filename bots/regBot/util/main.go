package util

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/tebeka/selenium"
)

const (
	LINUX_GECKO_PATH = "./bin/geckodriver"
	GECKO_PORT       = 5555
)

func LoggerInit(logID, descriptor string) *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	logger = slog.With(logID, descriptor)
	return logger
}

func ServerInit(port string, logger *slog.Logger) (string, error) {
	logger = logger.With("ServerInit", "Utilities")
	userOS, err := checkOS()
	if err != nil {
		logger.Error("Failed to check OS, Exiting.", "error", err)
		os.Exit(1)
	}
	logger.Info(userOS)
	status, err := fireWall(port, logger)
	if err != nil || !status {
		logger.Error("Error firewal returned false", "error", err)
		return "", err
	}
	return userOS, nil
}

func checkOS() (string, error) {
	os := runtime.GOOS
	if os == "darwin" {
		msg := "FUCK OFF"
		panic(msg)
	}

	return os, nil
}

func fireWall(port string, logger *slog.Logger) (bool, error) {
	logger = logger.With("FireWall", "Utilities")

	cmd := exec.Command("ufw", "allow", port)

	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("Failed to run ufw allow", "error", err)
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Printf("Output: %s\n", output)

	return true, nil
}

func BrowserProxyWindows(servicePort, workingProxy string, logger *slog.Logger) (selenium.WebDriver, error) {
	logger = logger.With("ID", "ApplyBrowserProxy")

	profileDir, err := os.MkdirTemp("", "firefox-profile")
	if err != nil {
		logger.Error("Failed to make temp dir")
	}
	defer os.RemoveAll(profileDir)

	parts := strings.Split(workingProxy, ":")
	proxyHost := parts[0]
	proxyPort := parts[1]
	// logger.Info("proxy parts HOST, "+proxyHost, "PORT:", proxyPort)
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

	encodedProfile := base64.StdEncoding.EncodeToString(buf.Bytes())

	caps := selenium.Capabilities{
		"browserName":     "firefox",
		"firefox_profile": encodedProfile,
	}
	wd, err := selenium.NewRemote(caps, servicePort)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to geckodriver: %w", err)
	}

	return wd, nil
}

func BrowserProxyLinux(binaryPath, driverPath, servicePort, workingProxy string, logger *slog.Logger) (selenium.WebDriver, *selenium.Service, error) {
	logger = logger.With("component", "BrowserProxyLinux")
	// Parse proxy
	parts := strings.Split(workingProxy, ":")
	if len(parts) != 2 {
		return nil, nil, fmt.Errorf("invalid proxy format: %s", workingProxy)
	}
	proxyHost := parts[0]
	proxyPort, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, nil, fmt.Errorf("invalid proxy port: %v", err)
	}
	logger.Info("proxy parts HOST, "+proxyHost, "PORT:", proxyPort)

	// Build Firefox preferences
	profileDir, err := os.MkdirTemp("", "firefox-profile")
	if err != nil {
		logger.Error("Failed to create temp Firefox profile dir", "error", err)
		return nil, nil, err
	}
	defer os.RemoveAll(profileDir)

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
		logger.Error("Failed to write prefs.js", "error", err)
		return nil, nil, err
	}

	// Zip and encode the profile
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)
	err = filepath.Walk(profileDir, func(path string, info os.FileInfo, err error) error {
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
	if err != nil {
		logger.Error("Failed to zip profile", "error", err)
		return nil, nil, err
	}
	zipWriter.Close()

	encodedProfile := base64.StdEncoding.EncodeToString(buf.Bytes())

	// Setup Firefox capabilities

	// logger.Info("using path", binaryPath, "for binary..")
	caps := selenium.Capabilities{
		"browserName": "firefox",
		"binary":      binaryPath,
		"moz:firefoxOptions": map[string]any{
			"args":    []string{"-headless"},
			"profile": encodedProfile,
		},
	}
	driverURL := "http://localhost:" + servicePort

	service, err := selenium.NewGeckoDriverService(LINUX_GECKO_PATH, GECKO_PORT)
	if err != nil {
		logger.Error("Failed to start geckodriver service", "error", err)
		return nil, nil, err
	}
	wd, err := selenium.NewRemote(caps, driverURL)
	if err != nil {
		logger.Error("Failed to create selenium WebDriver", "error", err)
		return nil, nil, err
	}
	return wd, service, nil
}
