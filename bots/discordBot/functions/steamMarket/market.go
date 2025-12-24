package steammarket

import (
	"discordBot/util"
	"io"
	"net/http"
	"net/url"
	"os"
)

const (
	STEAM_MARKET_API_URL = "steamapis.com/market/item{ AppID }/{ MarketHashName }?api_key={ YourSecretAPIKey }"
	BASE_API_URL         = "https://api.steamapis.com/market/item/730"
)

func marketHashName(param string) string {
	return url.QueryEscape(param)
}

func SteamItemAPICall(itemName string) (string, error) {
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
