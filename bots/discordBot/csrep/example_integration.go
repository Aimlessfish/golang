package main

import (
	"csrep/module"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

// Example of how to integrate the Steam Report Module into an existing Discord bot
func main() {
	// Load environment variables
	godotenv.Load()
	token := os.Getenv("DISCORD_BOT_TOKEN")

	// Create your existing Discord bot session
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal("Error creating Discord session:", err)
	}

	// Initialize the Steam Report Module with custom command prefix
	// Use "steam-" to avoid conflicts with existing bot commands
	steamModule := module.NewSteamReportModule(8080, "steam-")

	// Register the module's handlers with your bot
	steamModule.RegisterHandlers(dg)

	// You can also register other modules here
	// otherModule := module.NewOtherModule()
	// otherModule.RegisterHandlers(dg)

	// Set Discord intents
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages

	// Open the Discord connection
	err = dg.Open()
	if err != nil {
		log.Fatal("Error opening connection:", err)
	}
	defer dg.Close()

	log.Println("✅ Bot is running with Steam Report Module!")
	log.Printf("Steam module commands: %v", steamModule.GetCommands())

	// Wait for interrupt signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	// Shutdown modules gracefully
	steamModule.Shutdown()
	log.Println("Bot shutdown complete")
}
