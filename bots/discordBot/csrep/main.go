package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const startPort = 8080

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if len(os.Args) > 1 && os.Args[1] == "server" {
		runServer()
	} else {
		showUsage()
	}
}

func showUsage() {
	fmt.Println("Steam Bot Session Manager")
	fmt.Println("\nUsage:")
	fmt.Println("  csrep server    - Start the multi-session bot manager")
	fmt.Println("\nOnce running, use commands to manage sessions:")
	fmt.Println("  add <steam-username>  - Add a new session (auto-generates session ID)")
	fmt.Println("  list                  - List all sessions")
	fmt.Println("  remove <session-id>   - Remove a session")
	fmt.Println("  quit                  - Stop all sessions and exit")
}

func runServer() {
	manager := NewSessionManager(startPort)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down...")
		manager.StopAll()
		os.Exit(0)
	}()

	// Start control server for managing sessions
	go startControlServer(manager)

	fmt.Println("Steam Bot Session Manager started")
	fmt.Println("Control API running on :9000")
	fmt.Println("Sessions will start from port", startPort)
	fmt.Println("\nCommands:")
	fmt.Println("  add <steam-username>  - Add new session with browser login")
	fmt.Println("  list                  - List all sessions")
	fmt.Println("  remove <session-id>   - Remove a session")
	fmt.Println("  quit                  - Stop all and exit")
	fmt.Print("\n> ")

	// Interactive CLI
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			fmt.Print("> ")
			continue
		}

		parts := strings.Fields(line)
		cmd := parts[0]

		switch cmd {
		case "add":
			if len(parts) != 2 {
				fmt.Println("Usage: add <steam-username>")
			} else {
				userID := parts[1]
				fmt.Printf("\n⏳ Creating session for '%s'...\n", userID)
				fmt.Println("   A browser window will open momentarily.")
				fmt.Println("   Please complete the Steam login process.")
				fmt.Println("   This command will wait until login is complete.\n")

				session, err := manager.AddSession(userID)
				if err != nil {
					fmt.Printf("\n❌ Error: %v\n\n", err)
				} else {
					fmt.Printf("\n✅ Session '%s' is READY on port %d\n", session.ID, session.Port)
					fmt.Printf("   Bot is logged in and ready to receive commands!\n")
					fmt.Printf("\n   API Endpoints:\n")
					fmt.Printf("   - Status: http://localhost:%d/status\n", session.Port)
					fmt.Printf("   - Screenshot: http://localhost:%d/action/get-screenshot\n", session.Port)
					fmt.Printf("   - Open Profile: POST http://localhost:%d/action/open-profile\n", session.Port)
					fmt.Printf("   - Navigate: POST http://localhost:%d/action/navigate {\"url\":\"...\"}\n\n", session.Port)
				}
			}

		case "list":
			sessions := manager.ListSessions()
			if len(sessions) == 0 {
				fmt.Println("No active sessions")
			} else {
				fmt.Printf("\n%-15s %-20s %-10s %-12s\n", "SESSION ID", "USERNAME", "PORT", "LOGGED IN")
				fmt.Println(strings.Repeat("-", 62))
				for _, s := range sessions {
					loggedIn := "No"
					if s.isLoggedIn {
						loggedIn = "Yes"
					}
					fmt.Printf("%-15s %-20s %-10d %-12s\n", s.ID, s.UserID, s.Port, loggedIn)
				}
				fmt.Println()
			}

		case "remove":
			if len(parts) != 2 {
				fmt.Println("Usage: remove <session-id>")
			} else {
				sessionID := parts[1]
				err := manager.RemoveSession(sessionID)
				if err != nil {
					fmt.Printf("Error: %v\n", err)
				} else {
					fmt.Printf("Session '%s' removed\n", sessionID)
				}
			}

		case "quit", "exit":
			fmt.Println("Shutting down...")
			manager.StopAll()
			return

		default:
			fmt.Println("Unknown command. Available: add, list, remove, quit")
		}

		fmt.Print("> ")
	}
}

func startControlServer(manager *SessionManager) {
	mux := http.NewServeMux()

	mux.HandleFunc("/sessions", func(w http.ResponseWriter, r *http.Request) {
		sessions := manager.ListSessions()
		response := make([]map[string]interface{}, 0, len(sessions))
		for _, s := range sessions {
			response = append(response, map[string]interface{}{
				"id":     s.ID,
				"userId": s.UserID,
				"port":   s.Port,
			})
		}
		json.NewEncoder(w).Encode(response)
	})

	mux.HandleFunc("/sessions/add", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload struct {
			UserID string `json:"userId"`
		}

		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		session, err := manager.AddSession(payload.UserID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "success",
			"sessionId": session.ID,
			"port":      session.Port,
		})
	})

	log.Println("Control server starting on :9000")
	http.ListenAndServe(":9000", mux)
}
