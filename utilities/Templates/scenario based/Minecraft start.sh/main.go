// This is an example of template usage to deploy a PaperMC server start.sh file
// this script covers star.sh creation and template usage/manipulation

// runtime flags are -XMS -XMX -Threads

package main

import (
	"os"
	"text/template"
	"log/slog"
	"flag"
)

type PaperStartsh struct {
	XMS string
	XMX string
	Threads string
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	logger = slog.With("LogID", "PaperStartsh")

	xms := flag.String("XMS", "", "Minimum dedicated RAM")
	xmx := flag.String("XMX", "", "Maximum dedicated RAM")
	threads := flag.String("Threads", "", "Amount of dedicated paralleled CPU threads")

	flag.Parse()

	if *xms == "" || *xmx == "" || *threads == "" {
		logger.Info("-XMS -XMX -Threads values are required ")
		os.Exit(1)
	}

	// create slice of start.sh template
	config := []PaperStartsh{
		{
			XMS: *xms,
			XMX: *xmx,
			Threads: *threads,
		},
	}

	var tmplFile = "start.sh.tmpl"
	tmpl, err := template.ParseFiles(tmplFile)
	if err != nil {
		logger.Warn("Error, ", "Parsing start.sh.tmpl failed: ", err.Error())
		os.Exit(1)
	}
	file, err := os.Create("start.sh")
	if err != nil {
		logger.Warn("Error, ", "Creating start.sh failed: ", err.Error())
		os.Exit(1)
	}
	err = tmpl.Execute(file, config) //here would be the path to the server dir
	if err != nil {
		logger.Warn("Error, ", "Template execution failed", err.Error())
		os.Exit(1)
	}
	defer file.Close()
}