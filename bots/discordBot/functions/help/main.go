package help

import "github.com/bwmarrin/discordgo"

func DisplayHelp(channelID string, server *discordgo.Session, message *discordgo.MessageCreate) error {
	embeddedMsg := &discordgo.MessageEmbed{
		Title:       "Bot Help",
		Description: "All Bot Commands: ",
		Color:       0x00ffcc,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "!help", Value: "Displays this message!"},
			{Name: "!clear", Value: "Clears _100_ bot sent messages!"},
			{Name: "!football", Value: "Pulls all upcoming football matches in the UK and displays odds grouped by bookies and match"},
			{Name: "!proxy", Value: "Sends 1, tested; working, HTTP proxy."},
			{Name: "!number <length>", Value: "Generates a random number with the specified number of digits (1-18). Example: !number 5"},
			{Name: "!username <input>", Value: "Generates a realistic username based on your input. Example: !username JohnDoe"},
			{Name: "!string <length>", Value: "Generates a random string with the specified length. Example: !string 8"},
		},
	}
	server.ChannelMessageSendEmbed(channelID, embeddedMsg)
	return nil
}
