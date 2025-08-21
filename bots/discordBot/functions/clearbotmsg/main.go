package clearbotmsg

import (
	util "discordBot/util"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func ClearBotMessages(userID, channelID string, server *discordgo.Session, message *discordgo.MessageCreate) bool {
	logger := util.LoggerInit("ClearBotMessages", "clearbotmsg")

	limit := 100

	for {
		messages, err := server.ChannelMessages(message.ChannelID, limit, "", "", "")
		if err != nil {
			server.ChannelMessageSend(message.ChannelID, fmt.Sprintf("failed to load previous messages! %v ", err))
			return false
		}

		if len(messages) == 0 {
			return false
		}

		for _, msg := range messages {
			if msg.Author.ID != userID {
				v, err := util.MessageTTL(msg.ID)
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
		return true
	}

}
