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
		msg.Text = "You have subscribed to SSB rate updates."
	case "unsubscribe":
		msg.Text = "You have unsubscribed to SSB rate updates."
	case "rates":
		localTimezone, err := time.LoadLocation("Asia/Singapore") // Look up a location by it's IANA name.
		if err != nil {
			panic(err)
		}
		bonds, err := core.ListBonds(time.Now().In(localTimezone).AddDate(-1, 0, 0), time.Now().In(localTimezone), 12)
		if err != nil {
			log.Error(err)
			return
		}

		var bondReturns []float64
		var bondDates []string

		for _, bond := range *bonds {
			bondInterestRate, err := core.ListBondInterestRates(bond)
			if err != nil {
				log.Error(err)
				return
			}
			bondReturns = append(bondReturns, bondInterestRate.Year10Return)
			bondDates = append(bondDates, time.Time(bond.IssueDate).Format("Jan 06"))
		}
		buf, err := core.GenerateSSBInterestRatesChart(bondReturns, bondDates)
		if err != nil {
			log.Error(err)
			return
		}
		photoFileBytes := tgbotapi.FileBytes{
			Name:  "picture",
			Bytes: *buf,
		}
		photoConfig := tgbotapi.NewPhoto(update.Message.Chat.ID, photoFileBytes)
		photoConfig.Caption = "test message test test"
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
