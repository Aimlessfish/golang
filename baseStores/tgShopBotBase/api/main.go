package api

import (
	"log/slog"
	"os"

	"telegramconnect/handler"

	util "telegramconnect/util"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// THIS HANDLES INITAL USER COMMANDS
// /AND ROUTES THEM TO THE NECESSARY HANDLER
func CommandControl(bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	logger = logger.With("LogID", "CommandControl")
	switch message.Command() {
	case "start":
		handler.HandleStart(bot, message)
	case "help":
		handler.HandleHelp(bot, message)
	default:
		msg := tgbotapi.NewMessage(message.Chat.ID, "Unknown command - Use /help for help!")
		if _, err := bot.Send(msg); err != nil {
			logger.Warn("Error handling unknown command msg", "error", err.Error())
			return err
		}
	}
	return nil
}

// SELF EXPLANITORY
func HandleIncomingMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	// Check if the message is not nil
	if update.Message != nil {
		// If it's a command, process it with CommandControl
		if update.Message.IsCommand() {
			CommandControl(bot, update.Message)
		} else if update.CallbackQuery != nil { // Handle callback query if present
			go handler.HandleCallbackQuery(bot, update) // Call HandleCallbackQuery function
		}
	}
	return nil // Return nil if no errors
}

// CONNECT TO THE TELEGRAM API WITH YOUR API KEY
func ConnectAPI() (string, error) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	logger = logger.With("MAIN", "TG CONNECT")

	envVars, err := util.ValueGetter("TELEGRAM_BOT_TOKEN")
	if err != nil {
		logger.Warn("Error getting token", "Error", err.Error())
		return "Failed to get token", err
	}
	token := envVars["TELEGRAM_BOT_TOKEN"]
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		logger.Warn("Error running NewBot", "Error", err.Error())
		return "Failed to connect API key to TGAPI", err
	}
	logger.Info("Connected to bot " + bot.Self.UserName)

	update_channel := tgbotapi.NewUpdate(0)
	update_channel.Timeout = 60
	updates := bot.GetUpdatesChan(update_channel)

	for update := range updates {
		if update.Message != nil { //manage text
			logger.Info("Received message update", "chatID", update.Message.Chat.ID, "text", update.Message.Text)
			go HandleIncomingMessage(bot, update)
		} else if update.CallbackQuery != nil { //manage button presses
			logger.Info("Received callback query!", "callbackData", update.CallbackQuery.Data)
			go handler.HandleCallbackQuery(bot, update)
		}
	}

	return "", err
}
