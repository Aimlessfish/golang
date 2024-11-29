package main

import (
	"os"
	"log/slog"
	"net/http"
)

func downloader(url, filepath string) error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	logger = slog.With("logID", "Downloader")
	//Create file
	file, err := os.Create(filepath)
	if err != nil {
		logger.Warn(err.Error())
		os.Exit(1)
		return err
	}
	defer file.Close()

	resp, err := http.Get(url)
	if err != nil {
		logger.Warn(err.Error())
		os.Exit(1)
		return err
	}
	defer resp.Body.Close()

	//Inspect the servers response
	if resp.StatusCode != http.StatusOK {
		logger.Info("bad status: %v", resp.Status, err)
		os.Exit(1)
		return err
	}

	//Write body to the file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		logger.Warn(err.Error())
		os.Exit(1)
		return err
	}

	return nil
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	logger = slog.With("logID", "Main")

	url := "http://aimlessdev.co.uk/downloads/mods.zip" //minecraft modpack
	filePath:= "/home/Downloads/mods.zip" 
}