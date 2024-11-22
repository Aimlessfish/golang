// example in real world production of automated minecraft server deployment
// this script covers editing configs using go templates

// runtime flags are -Query -Rcon -Server

package main

import (
	"os"
	"fmt"
	"text/template"
	"log/slog"
	"flag"
)

type PaperDotProperties struct {
	QueryPort string
	RconPort string
	ServerPort string
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	logger = slog.With("LogID", "PaperDotProperties")

	// setup flags for the above structure
	Query := flag.String("QueryPort", "", "Server Query Port (int)")
	Rcon := flag.String("RconPort", "", "Server Rcon Port (int)")
	Server := flag.String("ServerPort", "", "Server port (int)")

	flag.Parse()

	if *Query == "" || *Rcon == "" || *Server == "" {
		logger.Info("-Query -Rcon -Server arguments are required")
		os.Exit(1)
	}


	// new slice of server.properties
	serverPorts := []PaperDotProperties{
		{
			QueryPort: *Query,
			RconPort: *Rcon,
			ServerPort: *Server,
		},
	}

	var tmplFile = "server.properties.tmpl"
	tmpl, err := template.ParseFiles(tmplFile)
	if err != nil {
		logger.Warn("Error", err.Error())
		os.Exit()
	}
	err = tmpl.Execute(os.Stdout, serverPorts)
	if err != nil {
		logger.Warn("Error,", "Template execution failed", err.Error())
		os.Exit(1)
	}

}