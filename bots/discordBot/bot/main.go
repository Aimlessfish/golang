package bot

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	clear "discordBot/functions/clearbotmsg"
	getproxy "discordBot/functions/proxy"
	util "discordBot/util"

	"github.com/bwmarrin/discordgo"
)

func ConnectAPI(logger *slog.Logger) error {
	logger = logger.With("Bot", "ConnectAPI")
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

	if message.Content == "list" || message.Content == "!list" || message.Content == "List" || message.Content == "/list" {
		server.ChannelMessageSend(message.ChannelID, "Not ready yet!")
		//logic to call all available public servers.
	}

	if message.Content == "!proxy" || message.Content == "Proxy" || message.Content == "http proxy" || message.Content == "proxy" || message.Content == "get http proxy" {
		proxies := getproxy.ProxyHandler(1)
		for _, proxy := range proxies {
			server.ChannelMessageSend(message.ChannelID, proxy)
		}
	}
	if message.Content == "http proxy list" || message.Content == "get http list" || message.Content == "proxy list" || message.Content == "list proxy" || message.Content == "proxies" {
		proxies := getproxy.ProxyHandler(0)
		for _, proxy := range proxies {
			server.ChannelMessageSend(message.ChannelID, proxy)
		}
	}
	if message.Content == "clear" {
		v := clear.ClearBotMessages(userID, channelID, server, message)
		if v {
			server.ChannelMessageSend(message.ChannelID, "cleared messages except this one lol")
		}
	}
}
