// Here is an example of creating directories with permissions
// requires elevated permissions to run

package main

import (
	"log/slog"
	"os"
)

func directories() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	logger = slog.With("LogID", "directories")

	path := "/home/example"
	err := os.MkdirAll(path, 0755)
	if err != nil {
		logger.Warn("Error running os.MkdirAll", "error", err.Error())
		os.Exit(1)
	}

}
