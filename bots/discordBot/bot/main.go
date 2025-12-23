package bot

import (
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	betting "discordBot/functions/betting"
	clear "discordBot/functions/clearbotmsg"
	"discordBot/functions/help"
	getproxy "discordBot/functions/proxy"
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

	return nil
}

func messageHandler(server *discordgo.Session, message *discordgo.MessageCreate) {
	userID := message.Author.ID
	channelID := message.ChannelID

	if message.Author.ID == server.State.User.ID {
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
		if !v {
			server.ChannelMessageSend(message.ChannelID, "failed to clear messages!")
		}
	}
	if message.Content == "!Football" || message.Content == "!football" {
		err := betting.MatchOdds(server, message)
		if err != nil {
			server.ChannelMessageSend(channelID, "Failed to retrieve upcoming matches! ")
		}
	}
	if strings.HasPrefix(message.Content, "!report") || strings.HasPrefix(message.Content, "report") {
		parts := util.SplitArgs(message.Content)
		if len(parts) < 2 {
			server.ChannelMessageSend(message.ChannelID, "Usage: !report <uid> <amount>")
			return
		}
		uid := parts[1]
		if uid == "" {
			server.ChannelMessageSend(message.ChannelID, "you did not provide a steam64 ID or a valid profile URL")
			return
		}
		if len(parts) >= 3 {
			amount := parts[2]
			server.ChannelMessageSend(message.ChannelID, amount+" Reports started for: \n (uid: "+uid+")")
			util.ExecReportBinary(uid, amount)
		} else {
			server.ChannelMessageSend(message.ChannelID, "Report started for: \n (uid: "+uid+")")
			util.ExecReportBinary(uid, "1")
		}
	}
}
