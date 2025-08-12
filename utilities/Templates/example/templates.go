// Basic example of using go.templates
// file manipulaiton and creation
// hardcoded filepath for ease

package main

import (
	"log/slog"
	"os"
	"text/template"
)

// Create the data structure
type Example struct {
	Name string
	Age  int
}

func main() {
	// new logger instance
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	logger = slog.With("LogID", "templates")

	// new slice of Example
	humans := []Example{
		{
			Name: "AimlessFish",
			Age:  20,
		},
		{
			Name: "Human2",
			Age:  15,
		},
	}

	var tmplFile = "resources/example.tmpl"
	tmpl, err := template.ParseFiles(tmplFile)
	if err != nil {
		logger.Warn("Error", err.Error())
		os.Exit(1)
	}
	err = tmpl.Execute(os.Stdout, humans)
	if err != nil {
		logger.Warn("Error", "Template execution", err.Error())
	}

}
