package bot

import (
	"log"

	"github.com/DanRulev/vocabot.git/internal/storage/cache"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ServiceI interface {
	WordSI
	QuizSI
}

type BotSender interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
}

type TelegramAPI struct {
	bot  *tgbotapi.BotAPI
	word *WordT
	quiz *QuizT
}

func NewTelegramAPI(botToken, env string, service ServiceI, cache *cache.Cache) (*TelegramAPI, error) {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, err
	}

	if env == "development" {
		bot.Debug = true
	} else {
		bot.Debug = false
	}

	return &TelegramAPI{
		bot:  bot,
		word: NewWordTAPI(bot, cache, service),
		quiz: NewQuizTAPI(bot, cache, service),
	}, nil
}

func (t *TelegramAPI) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := t.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			if update.Message.IsCommand() {
				t.handleCommand(update.Message)
			} else {
				t.handleMessage(update.Message)
			}
			continue
		}

		if update.CallbackQuery != nil {
			t.handleCallbackQuery(update.CallbackQuery)
		}
	}
}

func sendMessage(bot BotSender, msg tgbotapi.Chattable) {
	sentMsg, err := bot.Send(msg)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
	} else {
		log.Printf("Sent message to %d", sentMsg.Chat.ID)
	}
}
