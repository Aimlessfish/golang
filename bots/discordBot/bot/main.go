package bot

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	betting "discordBot/functions/betting"
	getproxy "discordBot/functions/proxy"
	initlogger "discordBot/util"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func getToken() string {
	logger := initlogger.LoggerInit("GET BOT", "BOT")
	logger.Info("Getting bot token")

	err := godotenv.Load()
	if err != nil {
		logger.Info("INVALID BOT TOKEN", "ERROR", err.Error())
		return "broke mate"
	} else {
		token := os.Getenv("TOKEN")
		if len(token) == 0 {
			panic("Token length == 0!")
		} else {
			return token
		}
	}
}

func ConnectAPI(logger *slog.Logger) error {
	logger = logger.With("Bot", "ConnectAPI")
	api_key := getToken()

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

	if message.Content == "!proxy" || message.Content == "Proxy" || message.Content == "http proxy" || message.Content == "proxy" || message.Content == "get http proxy" {
		proxies := getproxy.ProxyHandler(1)
		for _, proxy := range proxies {
			server.ChannelMessageSend(message.ChannelID, proxy)
		}
	}

	if message.Content == "clear" {
		var msgmap []string
		limit := 100

		for {
			messages, err := server.ChannelMessages(channelID, limit, "", "", "")
			if err != nil {
				server.ChannelMessageSend(channelID, fmt.Sprintf("failed to load previous messages! %v ", err))
				break
			}

			if len(messages) == 0 {
				break
			}

			for _, msg := range messages {
				if msg.Author.ID == userID || msg.Author.ID == server.State.User.ID {
					//logic here for working out 14 day limit

					//if >14 days print("delete manually")

				}

			}
		}

		err := server.ChannelMessagesBulkDelete(message.ChannelID, msgmap)
		if err != nil {
			server.ChannelMessageSend(channelID, fmt.Sprintf("Failed %v", err))
		}
	}

	if message.Content == "!Football" || message.Content == "!football" {
		err := betting.MatchOdds(server, message)
		if err != nil {
			server.ChannelMessageSend(channelID, "Failed to retrieve upcoming matches! ")
		}
	}
}
