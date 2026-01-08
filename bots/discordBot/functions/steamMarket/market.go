package steammarket

import (
	"discordBot/util"
	"encoding/json"
	"io"
	"net/http"
)

var data steamMarketInjection

type steamMarketInjection struct {
	Success     bool   `json:"success"`
	PriceLow    string `json:"lowest_price"`
	Volume      string `json:"volume"`
	PriceMedian string `json:"median_price"`
}

func SteamMarketPriceOverviewURL(itemName string) (string, error) {
	logger := util.LoggerInit("UTIL", "SteamMarketPriceOverviewURL")

	// Ensure item name is URL-escaped
	itemHash := util.MarketHashName(itemName)
	output, err := util.FormatForSteamMarketInjection(itemHash)
	if err != nil && output == "" {
		logger.Error("URL formatting failed", "error", err, "output", output)
		return "URL formatting failed", err
	}
	response, err := http.Get(output)
	if err != nil {
		logger.Error("Failed to get request", "error", err, "url", output)
		return "HTTP request failed", err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		logger.Error("Failed to read response body", "error", err)
		return "Read body failed", err
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		logger.Error("Failed to parse Steam Web API JSON", "error", err, "body", string(body))
		return "JSON parse failed", err
	}
	if !data.Success {
		logger.Error("Steam API did not return success", "body", string(body))
		return "Steam API did not return success", nil
	}
	result := "Price - Low: " + data.PriceLow + "\n" +
		"Median: " + data.PriceMedian
	return result, nil

}
