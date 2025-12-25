package bot

import (
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	betting "discordBot/functions/betting"
	clear "discordBot/functions/clearbotmsg"
	"discordBot/functions/generators"
	"discordBot/functions/help"
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

	if strings.HasPrefix(message.Content, "!number ") {
		parts := strings.Split(message.Content, " ")
		if len(parts) != 2 {
			server.ChannelMessageSend(channelID, "Please provide a valid length for the random number. Example: !number 5")
			return
		}
		input := parts[1]
		length, err := strconv.Atoi(input)
		if err != nil || length <= 0 {
			server.ChannelMessageSend(channelID, "Please provide a valid positive integer for the length.")
			return
		}
		if length > 18 {
			server.ChannelMessageSend(channelID, "Please provide a number length between 1 and 18.")
			return
		}
		randomNumber := generators.GenerateRandomNumber(input)
		server.ChannelMessageSend(channelID, "Generated Random Number: "+randomNumber)
	}

	if strings.HasPrefix(message.Content, "!username ") {
		parts := strings.Split(message.Content, " ")
		if len(parts) != 2 {
			server.ChannelMessageSend(channelID, "Please provide a valid input for the username. Example: !username JohnDoe")
			return
		}
		input := parts[1]
		username := generators.GenerateUsername(input)
		server.ChannelMessageSend(channelID, "Generated Username: "+username)
	}
	if strings.HasPrefix(message.Content, "!string") {
		parts := strings.Split(message.Content, " ")
		if len(parts) != 2 {
			server.ChannelMessageSend(channelID, "Please provide a valid length for the random string. Example: !string 10")
			return
		}
		length, err := strconv.Atoi(parts[1])
		if err != nil || length <= 0 {
			server.ChannelMessageSend(channelID, "Please provide a valid positive integer for the length.")
			return
		}
		randomString := generators.GenerateRandomString(length)
		server.ChannelMessageSend(channelID, "Generated Random String: "+randomString)
	}
}
