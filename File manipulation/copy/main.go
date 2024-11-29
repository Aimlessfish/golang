package main

import(
	"log/slog"
	"io"
	"fmt"
)

func copyFile(src string, dest string) error{
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	logger = slog.With("logID", "copyFile")
	
	bytesRead, err := ioutil.ReadFile(src)
	if err != nil {
		logger.Warn("Error reading source file", err.Error())
		return err
	}

	err = ioutil.WriteFile(dest, bytesRead, 0644)
	if err != nil {
		logger.Warn("Error writing destination file", err.Error())
		return err
	}
	
}

func main(){
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	logger = slog.With("logID", "Main")

	source := flag.String("Source", "", "Source File Location")
	destination := flag.String("Destination", "", "Destination")
	flag.Parse()
	
	err := copyFile(source, destination)
	if err != nil {
		logger.Warn("Error running copyFile", err.Error())
	}
}
