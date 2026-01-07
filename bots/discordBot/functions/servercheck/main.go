package servercheck

import (
	"encoding/json"
	"fmt"

	"discordBot/util"
)

// Server represents a game server to check
type Server struct {
	Name string
	Port int
}

// ServerResponse represents the JSON structure returned by the serviceChecker binary
type ServerResponse struct {
	Name    string
	Port    int
	Address string `json:"address"`
	Status  string `json:"status"`
	Message string
}

// CheckServers calls the serviceChecker binary for each server, parses the JSON responses, and returns a user-friendly string
func CheckServers() (string, error) {
	ip := "65.21.132.169"
	servers := []Server{
		{"Minecraft Vanilla", 25565},
		{"Minecraft Modded: Society Sunlit Valley", 25566},
		{"Valheim", 24567},
		// Add more servers as needed
	}

	var responses []ServerResponse

	for _, srv := range servers {
		args := fmt.Sprintf("%s:%d", ip, srv.Port)
		output, err := util.ExecBinary("./bin/serviceChecker", args)
		if err != nil {
			responses = append(responses, ServerResponse{
				Name:    srv.Name,
				Port:    srv.Port,
				Status:  "Error",
				Message: fmt.Sprintf("Failed to check: %s", err.Error()),
			})
			continue
		}

		var response ServerResponse
		err = json.Unmarshal([]byte(output), &response)
		if err != nil {
			responses = append(responses, ServerResponse{
				Name:    srv.Name,
				Port:    srv.Port,
				Status:  "Error",
				Message: fmt.Sprintf("Failed to parse: %s\nRaw: %s", err.Error(), output),
			})
			continue
		}

		// Ensure the response has the correct name and port
		response.Name = srv.Name
		response.Port = srv.Port
		responses = append(responses, response)
	}

	// Format the response in a user-friendly way
	if len(responses) == 0 {
		return "🔍 No servers to check.", nil
	}

	result := "🔍 Server Check Results:\n"
	for _, resp := range responses {
		result += fmt.Sprintf("**%s** (Port %d):\n  Status: %s\n  Details: %s\n", resp.Name, resp.Port, resp.Status, resp.Message)
	}
	return result, nil
}
