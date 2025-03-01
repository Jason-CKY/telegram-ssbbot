package handler

import (
	"github.com/Jason-CKY/telegram-ssbbot/pkg/schemas"
	"github.com/Jason-CKY/telegram-ssbbot/pkg/utils"
	log "github.com/sirupsen/logrus"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleUpdate(update *tgbotapi.Update, bot *tgbotapi.BotAPI) {
	if update.Message != nil {
		chatSettings, _, err := schemas.InsertChatSettingsIfNotPresent(update.Message.Chat.ID)
		if err != nil {
			log.Error(err)
			return
		}
		if update.Message.IsCommand() {
			HandleCommand(update, bot, chatSettings)
		}
	}
}

func HandleCommand(update *tgbotapi.Update, bot *tgbotapi.BotAPI, chatSettings *schemas.ChatSettings) {
	// Create a new MessageConfig. We don't have text yet,
	// so we leave it empty.
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	// Extract the command from the Message.
	switch update.Message.Command() {
	case "help":
		msg.Text = utils.HELP_MESSAGE
	case "subscribe":
		msg.Text = "subscribing"
	default:
		return
	}

	if _, err := bot.Request(msg); err != nil {
		log.Error(err)
		return
	}
}
