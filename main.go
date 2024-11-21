// Example of using argument flags in Go.
// Get, parse, output parse.

// Example of structured logging also included.

package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	logger = slog.With("logID", "Main")
	// Get
	flag1 := flag.String("flag1", "", "string")
	flag2 := flag.Int("flag2", 0, "integer")

	// Parse
	flag.Parse()
	logger.Info("Flags parsed successfully..")

	if *flag1 == "" || *flag2 == 0 {
		logger.Warn("Flags empty,", "error", "-flag1 and -flag2 are required (string, int) ")
		os.Exit(1)
	}

	// Output
	fmt.Print("flag1 = ", *flag1, " flag2 = ", *flag2)
}
