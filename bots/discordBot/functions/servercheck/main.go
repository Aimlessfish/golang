package servercheck

import (
	"encoding/json"
	"fmt"

	"discordBot/util"
)

// ServerResponse represents the JSON structure returned by the serviceChecker binary
type ServerResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	// Add more fields as needed based on the actual JSON output
}

// CheckServer calls the serviceChecker binary, parses the JSON response, and returns a user-friendly string
func CheckServer() (string, error) {
	// Assuming the binary is located at ./bin/serviceChecker and takes no arguments
	output, err := util.ExecBinary("./bin/serviceChecker", "")
	if err != nil {
		return fmt.Sprintf("❌ Failed to check server status: %s", err.Error()), err
	}

	var response ServerResponse
	err = json.Unmarshal([]byte(output), &response)
	if err != nil {
		return fmt.Sprintf("❌ Failed to parse server response: %s\nRaw output: %s", err.Error(), output), err
	}

	// Format the response in a user-friendly way
	return fmt.Sprintf("🔍 Server Check Result:\nStatus: %s\nDetails: %s", response.Status, response.Message), nil
}
