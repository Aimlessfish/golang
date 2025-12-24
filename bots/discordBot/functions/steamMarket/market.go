package steammarket

import (
	"discordBot/util"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

const (
	STEAM_MARKET_API_URL = "steamapis.com/market/item{ AppID }/{ MarketHashName }?api_key={ YourSecretAPIKey }"
	BASE_API_URL         = "https://api.steamapis.com/market/item/730"
	STEAM_WEB_API_URL    = "https://www.steamwebapi.com/steam/api/item?key="
	EMPIRE_API_URL       = " https://api.pricempire.com/v4/paid/items/prices?app_id=730"
)

func marketHashName(param string) string {
	return url.QueryEscape(param)
}

func SteamWebApiItemCall(itemName string) (string, error) {
	logger := util.LoggerInit("UTIL", "SteamWebAPICall")

	// get the market hash name
	hash := marketHashName(itemName)
	resp, err := http.Get(STEAM_WEB_API_URL + os.Getenv("STEAMWEB_API_KEY") + "&market_hash_name=" + hash + "&currency=GBP")
	if err != nil {
		logger.Error("Failed to make Steam Web API call", "error", err)
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logger.Error("Steam Web API call returned non-200 status", "status", resp.StatusCode)
		return "", err
	}

	// Parse the JSON response
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read Steam Web API response body", "error", err)
		return "", err
	}

	type steamMarketResponse struct {
		PriceLatestSell float64 `json:"pricelatestsell"`
		PriceMedian     float64 `json:"pricemedian"`
		PriceMin        float64 `json:"pricemin"`
		Sold24h         int     `json:"sold24h"`
	}

	var data steamMarketResponse
	err = json.Unmarshal(bodyBytes, &data)
	if err != nil {
		logger.Error("Failed to parse Steam Web API JSON", "error", err)
		return "", err
	}

	result := "pricelatestsell: " + fmt.Sprintf("%.2f", data.PriceLatestSell) + "\n" +
		"pricemedian: " + fmt.Sprintf("%.2f", data.PriceMedian) + "\n" +
		"pricemin: " + fmt.Sprintf("%.2f", data.PriceMin) + "\n" +
		"sold24h: " + fmt.Sprintf("%d", data.Sold24h)

	return result, nil
}

// Read and return the response body as a string

func SteamApisItemCall(itemName string) (string, error) {
	logger := util.LoggerInit("UTIL", "SteamMarketAPICall")

	//get the market hash name
	hash := marketHashName(itemName)
	resp, err := http.Get(BASE_API_URL + "/" + hash + "?api_key=" + os.Getenv("STEAM_API_KEY"))
	if err != nil {
		logger.Error("Failed to make Steam Market API call", "error", err)
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logger.Error("Steam Market API call returned non-200 status", "status", resp.StatusCode)
		return "", err
	}

	// Read and return the response body as a string
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read Steam Market API response body", "error", err)
		return "", err
	}
	return string(bodyBytes), nil
}

func PriceEmpireItemCall(itemName string) (string, error) {
	logger := util.LoggerInit("UTIL", "PriceEmpireItemCall")
	apiKey := os.Getenv("PRICEEMPIRE_API_KEY")
	if apiKey == "" {
		logger.Error("PRICEEMPIRE_API_KEY not set")
		return "", fmt.Errorf("PRICEEMPIRE_API_KEY not set")
	}
	req, err := http.NewRequest("GET", "https://api.pricempire.com/v4/free/search?q="+itemName, nil)
	if err != nil {
		logger.Error("Failed to create Price Empire request", "error", err)
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Failed to make Price Empire API call", "error", err)
		return "", err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logger.Error("Price Empire API call returned non-200 status", "status", resp.StatusCode)
		return "", fmt.Errorf("Price Empire API call returned status %d", resp.StatusCode)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read Price Empire API response body", "error", err)
		return "", err
	}
	return string(bodyBytes), nil
}
