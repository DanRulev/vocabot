package mock_bot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type MockBot struct {
	SentMessages []tgbotapi.Chattable
}

func (m *MockBot) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	m.SentMessages = append(m.SentMessages, c)
	return tgbotapi.Message{Chat: &tgbotapi.Chat{ID: 123}}, nil
}

func ClearSentMessages(bot *MockBot) {
	bot.SentMessages = nil
}
