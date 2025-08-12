// example in real world production of automated minecraft server deployment
// this script covers editing configs using go templates

// runtime flags are -QueryPort -RconPort -ServerPort

package main

import (
	"os"
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
		logger.Info("-QueryPort -RconPort -ServerPort arguments are required")
		os.Exit(1)
	}


	// new slice of server.properties template
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
		logger.Warn("Error, ", "Parsing tempate failed: ", err.Error())
		os.Exit(1)
	}
	file, err := os.Create("server.properties")
	if err != nil {
		logger.Warn("Fatal", "Could not create server.properties", err.Error())
		os.Exit(1)
	}
	err = tmpl.Execute(file, serverPorts) //here would be the path to the server dir you are setting up
	if err != nil {
		logger.Warn("Error,", "Template execution failed", err.Error())
		os.Exit(1)
	}
	defer file.Close()

}