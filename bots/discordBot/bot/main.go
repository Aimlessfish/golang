package bot

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	betting "discordBot/functions/betting"
	clear "discordBot/functions/clearbotmsg"
	"discordBot/functions/help"
	getproxy "discordBot/functions/proxy"
	"discordBot/functions/steam"
	util "discordBot/util"

	"github.com/bwmarrin/discordgo"
)

func ConnectAPI(logger *slog.Logger) error {
	logger = logger.With("Bot", "ConnectAPI")

	// Load environment variables
	if err := util.LoadEnv(); err != nil {
		logger.Warn("No .env file found, using defaults")
	}

	api_key := util.GetToken()

	// Initialize Steam session manager with port from env
	startPort := util.GetEnvAsInt("SESSION_START_PORT", 8080)
	steam.InitSteamManager(startPort)

	logger.Info("Steam manager initialized", "start_port", startPort)

	discord, err := discordgo.New("Bot " + api_key)
	if err != nil {
		logger.Error("API connect failed!")
		return err
	}
	err = discord.Open()
	if err != nil {
		panic(err)
	}
	discord.AddHandler(messageHandler)
	logger.Info("Bot is running. Press CTRL+C to exit.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop

	// Cleanup Steam sessions on shutdown
	if manager := steam.GetManager(); manager != nil {
		manager.StopAll()
	}

	return nil
}

func messageHandler(server *discordgo.Session, message *discordgo.MessageCreate) {
	userID := message.Author.ID
	channelID := message.ChannelID

	if message.Author.ID == server.State.User.ID {
		return
	}

	// Check for Steam commands first (they handle their own loading messages)
	if steam.HandleSteamCommands(server, message) {
		return
	}

	server.ChannelMessageSendReply(message.ChannelID, "loading..", &discordgo.MessageReference{
		MessageID: message.ID,
		ChannelID: message.ChannelID,
		GuildID:   message.GuildID,
	})

	if message.Content == "!help" || message.Content == "help" || message.Content == "commands" {
		help.DisplayHelp(channelID, server, message)
	}
	if message.Content == "!proxy" || message.Content == "Proxy" || message.Content == "http proxy" || message.Content == "proxy" || message.Content == "get http proxy" {
		proxies := getproxy.ProxyHandler(1)
		for _, proxy := range proxies {
			server.ChannelMessageSend(message.ChannelID, proxy)
		}
	}
	if message.Content == "!clear" {
		v := clear.ClearBotMessages(userID, channelID, server, message)
		if v {
			server.ChannelMessageSend(message.ChannelID, "cleared messages except this one lol")
		}
	}

	if message.Content == "!Football" || message.Content == "!football" {
		err := betting.MatchOdds(server, message)
		if err != nil {
			server.ChannelMessageSend(channelID, "Failed to retrieve upcoming matches! ")
		}
	}
}
