package bot

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	getproxy "discordBot/functions/proxy"
	util "discordBot/util"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func getToken() string {
	logger := util.LoggerInit("GET BOT", "BOT")
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
	logger := util.LoggerInit("messageHandler", "commands")
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
				if msg.Author.ID != userID {
					v, err := messageTTL(msg.ID)
					if !v || err != nil {
						logger.Error("TTL expired", "error", err)
						continue
					} else {
						err = server.ChannelMessageDelete(message.ChannelID, msg.ID)
						if err != nil {
							continue
						}
					}
				} else {
					continue
				}

			}

		}
	}
}

func messageTTL(msgID string) (bool, error) {
	logger := util.LoggerInit("MAIN", "messageTLL")
	const discordEpoch = 1420070400000

	id64, err := strconv.ParseInt(msgID, 10, 64)
	if err != nil {
		logger.Error("Failed to parse Message Date from msg.ID", "error", err)
		os.Exit(1)
	}

	timestamp := (id64 >> 22) + discordEpoch
	messageTime := time.UnixMilli(timestamp)

	if time.Since(messageTime) > (14 * 24 * time.Hour) {
		return false, nil
	}

	return true, nil
}
