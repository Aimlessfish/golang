package util

import (
	"runtime"
)

func CheckOS() (string, error) {
	os := runtime.GOOS
	if os == "darwin" {
		msg := "FUCK OFF"
		panic(msg)
	}

	return os, nil
}

func BrowserInit() error {

	return nil
}
