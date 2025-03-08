package handler

import (
	"time"

	"github.com/Jason-CKY/telegram-ssbbot/pkg/core"
	"github.com/Jason-CKY/telegram-ssbbot/pkg/schemas"
	"github.com/Jason-CKY/telegram-ssbbot/pkg/utils"
	log "github.com/sirupsen/logrus"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleUpdate(update *tgbotapi.Update, bot *tgbotapi.BotAPI) {
	if update.Message != nil && utils.IsUsernameAllowed(update.Message.From.UserName) {
		if update.Message.IsCommand() {
			HandleCommand(update, bot)
		}
	}
}

func HandleCommand(update *tgbotapi.Update, bot *tgbotapi.BotAPI) {
	// Create a new MessageConfig. We don't have text yet,
	// so we leave it empty.
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

	// Extract the command from the Message.
	switch update.Message.Command() {
	case "help":
		msg.Text = utils.HELP_MESSAGE
	case "subscribe":
		_, _, err := schemas.InsertChatSettingsIfNotPresent(update.Message.Chat.ID)
		if err != nil {
			log.Error(err)
			return
		}
		msg.Text = "You have subscribed to SSB rate updates."
	case "unsubscribe":
		chatSettings, _, err := schemas.InsertChatSettingsIfNotPresent(update.Message.Chat.ID)
		if err != nil {
			log.Error(err)
			return
		}
		err = chatSettings.Delete()
		if err != nil {
			log.Error(err)
			return
		}
		msg.Text = "You have unsubscribed to SSB rate updates."
	case "rates":
		localTimezone, err := time.LoadLocation("Asia/Singapore") // Look up a location by it's IANA name.
		if err != nil {
			log.Error(err)
			return
		}
		photoConfig, err := core.GenerateNotificationMessage(update.Message.Chat.ID, localTimezone)
		if err != nil {
			log.Error(err)
			return
		}
		if _, err := bot.Send(photoConfig); err != nil {
			log.Error(err)
			return
		}
		return
	default:
		return
	}

	if _, err := bot.Request(msg); err != nil {
		log.Error(err)
		return
	}
}
