// Example of commandline command creation and execution
// This can be useful for executing system services not inlcuded -
// in the go standard package library

// This example is of an automated mail query

package main

import (
	"log/slog"
	"os"
	"os/exec"
)

func main() {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	logger = slog.With("LogID", "command creation")

	cmd := exec.Command("sh", "-c", "echo helloworld | mail 'example@example.com'")
	err := cmd.Run()
	if err != nil {
		logger.Warn("Error running command", "error", err.Error())
		os.Exit(1)
	}

}
