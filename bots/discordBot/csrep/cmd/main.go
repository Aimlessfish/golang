package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"csrep/bot"

	"github.com/joho/godotenv"
)

var (
	startPort       int
	defaultTimeout  int
	controlAPIPort  int
	discordToken    string
	headlessBrowser bool
)

func init() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using defaults")
	}

	// Load configuration from environment variables
	startPort = getEnvAsInt("SESSION_START_PORT", 8080)
	defaultTimeout = getEnvAsInt("SESSION_TIMEOUT_MINUTES", 0)
	controlAPIPort = getEnvAsInt("CONTROL_API_PORT", 9000)
	discordToken = os.Getenv("DISCORD_BOT_TOKEN")
	headlessBrowser = getEnvAsBool("HEADLESS_BROWSER", false)
}

func getEnvAsInt(key string, defaultVal int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultVal
}

func getEnvAsBool(key string, defaultVal bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultVal
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "server":
			runCLIServer()
		case "discord":
			runDiscordBot()
		default:
			showUsage()
		}
	} else {
		showUsage()
	}
}

func showUsage() {
	fmt.Println("Steam Bot Session Manager")
	fmt.Println("\nUsage:")
	fmt.Println("  csrep server    - Start the CLI session manager")
	fmt.Println("  csrep discord   - Start the Discord bot")
	fmt.Println("\nConfiguration:")
	fmt.Println("  Create a .env file (see .env.example) to configure:")
	fmt.Println("  - DISCORD_BOT_TOKEN: Your Discord bot token")
	fmt.Println("  - SESSION_START_PORT: Starting port for sessions (default: 8080)")
	fmt.Println("  - SESSION_TIMEOUT_MINUTES: Default timeout in minutes (default: 0/no timeout)")
	fmt.Println("  - CONTROL_API_PORT: Control API port (default: 9000)")
	fmt.Println("  - HEADLESS_BROWSER: Run browser headless (default: false)")
	fmt.Println("\nCLI Commands (when in server mode):")
	fmt.Println("  add <steam-username> [timeout]  - Add session (optional timeout in minutes)")
	fmt.Println("  list                            - List all sessions")
	fmt.Println("  remove <session-id>             - Remove a session")
	fmt.Println("  extend <session-id> <minutes>   - Extend session timeout")
	fmt.Println("  quit                            - Stop all sessions and exit")
}

func runCLIServer() {
	manager := bot.NewSessionManager(startPort)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down...")
		manager.StopAll()
		os.Exit(0)
	}()

	fmt.Println("Steam Bot Session Manager started")
	fmt.Println("Sessions will start from port", startPort)
	if defaultTimeout > 0 {
		fmt.Printf("Default session timeout: %d minutes\n", defaultTimeout)
	}
	fmt.Println("\nCommands:")
	fmt.Println("  add <steam-username> [timeout-minutes]")
	fmt.Println("  list")
	fmt.Println("  remove <session-id>")
	fmt.Println("  extend <session-id> <minutes>")
	fmt.Println("  cred-add <account> <password|qr> [shared_secret]")
	fmt.Println("  cred-list")
	fmt.Println("  cred-remove <account>")
	fmt.Println("  quit")
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
			if len(parts) < 2 {
				fmt.Println("Usage: add <steam-username> [timeout-minutes]")
			} else {
				userID := parts[1]
				var timeout time.Duration

				fmt.Printf("\n⏳ Creating session for '%s'...\n", userID)
				fmt.Println("   A browser window will open momentarily.")
				fmt.Println("   Please complete the Steam login process.")
				fmt.Println("   This command will wait until login is complete.")

				var session *bot.SessionChrome
				var err error

				if len(parts) >= 3 {
					// Parse timeout from command
					var minutes int
					fmt.Sscanf(parts[2], "%d", &minutes)
					timeout = time.Duration(minutes) * time.Minute
					fmt.Printf("   Auto-remove after %d minutes\n\n", minutes)
					session, err = manager.AddSessionWithTimeout(userID, timeout)
				} else if defaultTimeout > 0 {
					// Use default timeout from .env
					timeout = time.Duration(defaultTimeout) * time.Minute
					fmt.Printf("   Auto-remove after %d minutes (default)\n\n", defaultTimeout)
					session, err = manager.AddSessionWithTimeout(userID, timeout)
				} else {
					session, err = manager.AddSession(userID)
				}

				if err != nil {
					fmt.Printf("\n❌ Error: %v\n\n", err)
				} else {
					fmt.Printf("\n✅ Session '%s' is READY on port %d\n", session.ID, session.Port)
					fmt.Printf("   Bot is logged in and ready to receive commands!\n")
					if timeout > 0 {
						fmt.Printf("   Will auto-remove in %v\n", timeout)
					}
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
				fmt.Printf("\n%-15s %-20s %-10s %-12s %-20s\n", "SESSION ID", "USERNAME", "PORT", "LOGGED IN", "CREATED")
				fmt.Println(strings.Repeat("-", 82))
				for _, s := range sessions {
					loggedIn := "No"
					if s.IsLoggedIn {
						loggedIn = "Yes"
					}
					fmt.Printf("%-15s %-20s %-10d %-12s %-20s\n",
						s.ID, s.UserID, s.Port, loggedIn, s.CreatedAt.Format("15:04:05"))
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

		case "extend":
			if len(parts) != 3 {
				fmt.Println("Usage: extend <session-id> <minutes>")
			} else {
				sessionID := parts[1]
				var minutes int
				fmt.Sscanf(parts[2], "%d", &minutes)
				duration := time.Duration(minutes) * time.Minute

				err := manager.ExtendTimeout(sessionID, duration)
				if err != nil {
					fmt.Printf("Error: %v\n", err)
				} else {
					fmt.Printf("Session '%s' timeout extended by %d minutes\n", sessionID, minutes)
				}
			}

		case "quit", "exit":
			fmt.Println("Shutting down...")
			manager.StopAll()
			return

		case "cred-add":
			if len(parts) < 3 {
				fmt.Println("Usage: cred-add <account> <password|qr> [shared_secret]")
				fmt.Println("       cred-add <email> <account> <password> [shared_secret]  (bot account)")
				fmt.Println("  Login methods:")
				fmt.Println("    - password: cred-add myaccount mypassword [shared_secret]")
				fmt.Println("    - bot:      cred-add email@example.com botname botpass [shared_secret]")
				fmt.Println("    - QR code:  cred-add myaccount qr")
			} else {
				var cred *bot.SteamCredential

				// Check if this is a bot account (4+ args or email pattern)
				if len(parts) >= 4 && strings.Contains(parts[1], "@") {
					// Bot account format: email username password [shared_secret]
					email := parts[1]
					account := parts[2]
					password := parts[3]
					sharedSecret := ""
					if len(parts) >= 5 {
						sharedSecret = parts[4]
					}

					cred = &bot.SteamCredential{
						AccountName:  account,
						Email:        email,
						Password:     password,
						SharedSecret: sharedSecret,
						LoginMethod:  bot.LoginMethodPassword,
					}
					fmt.Printf("Adding bot account: %s (%s)\n", account, email)
				} else {
					// Original format
					account := parts[1]
					passwordOrMethod := parts[2]

					// Check if user wants QR login
					if passwordOrMethod == "qr" {
						cred = &bot.SteamCredential{
							AccountName: account,
							LoginMethod: bot.LoginMethodQR,
						}
						fmt.Printf("Setting up QR code login for account: %s\n", account)
					} else {
						// Username/password login
						password := passwordOrMethod
						sharedSecret := ""
						if len(parts) >= 4 {
							sharedSecret = parts[3]
						}

						cred = &bot.SteamCredential{
							AccountName:  account,
							Password:     password,
							SharedSecret: sharedSecret,
							LoginMethod:  bot.LoginMethodPassword,
						}
					}
				}

				err := manager.AddCredential(cred)
				if err != nil {
					fmt.Printf("Error: %v\n", err)
				} else {
					fmt.Printf("✓ Credentials saved for account: %s\n", cred.AccountName)
					if cred.Email != "" {
						fmt.Printf("  Email: %s\n", cred.Email)
					}
					if cred.LoginMethod == bot.LoginMethodQR {
						fmt.Println("  Login method: QR code")
						fmt.Println("  You'll scan a QR code with your Steam Mobile app when logging in")
					} else {
						fmt.Println("  Login method: Username + Password")
						if cred.SharedSecret != "" {
							fmt.Println("  (includes Steam Guard shared secret)")
						}
					}
				}
			}

		case "cred-list":
			accounts := manager.ListCredentials()
			if len(accounts) == 0 {
				fmt.Println("No credentials stored")
			} else {
				fmt.Println("\nStored Steam Accounts:")
				for _, acc := range accounts {
					fmt.Printf("  - %s\n", acc)
				}
				fmt.Println()
			}

		case "cred-remove":
			if len(parts) != 2 {
				fmt.Println("Usage: cred-remove <account>")
			} else {
				account := parts[1]
				err := manager.RemoveCredential(account)
				if err != nil {
					fmt.Printf("Error: %v\n", err)
				} else {
					fmt.Printf("✓ Credentials removed for: %s\n", account)
				}
			}

		default:
			fmt.Println("Unknown command. Available: add, list, remove, extend, cred-add, cred-list, cred-remove, quit")
		}

		fmt.Print("> ")
	}
}

func runDiscordBot() {
	if discordToken == "" {
		fmt.Println("❌ Error: DISCORD_BOT_TOKEN not set")
		fmt.Println("\nPlease create a .env file with your Discord bot token:")
		fmt.Println("  DISCORD_BOT_TOKEN=your_token_here")
		fmt.Println("\nSee .env.example for full configuration options")
		os.Exit(1)
	}

	bot := NewDiscordBot(discordToken)
	if err := bot.Start(); err != nil {
		log.Fatalf("Failed to start Discord bot: %v", err)
	}
}
