package util

import (
	"runtime"
)

func CheckOS() (string, error) {
	os := runtime.GOOS
	return os, nil
}

func BrowserInit() error {

	return nil
}
