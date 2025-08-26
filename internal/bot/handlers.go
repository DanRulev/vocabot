package bot

import (
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	ButtonNewWord         = "📚 Новое слово"
	ButtonQuiz            = "🧠 Викторина"
	ButtonMyWords         = "❗Мои слова"
	ButtonProgress        = "📊 Мой прогресс"
	ButtonWordProgress    = "📚 Слова"
	ButtonQuizProgress    = "🧠 Викторины"
	ButtonLearnedWords    = "✅ Выученные слова"
	ButtonNotLearnedWords = "❌ Не выученные слова"
	ButtonMainMenu        = "🏠 Главное меню"
	ButtonBack            = "⏪ Назад"
	ButtonHelp            = "ℹ️ Помощь"
)

func (t *TelegramAPI) handleCommand(message *tgbotapi.Message) {
	switch message.Command() {
	case "start":
		t.handleStartCommand(message)
	case "help":
		t.handleHelpCommand(message)
	default:
		msg := tgbotapi.NewMessage(message.Chat.ID, "Неизвестная команда. Используй /start")
		sendMessage(t.bot, msg)
	}
}

func (t *TelegramAPI) handleStartCommand(message *tgbotapi.Message) {
	welcomeText := "🤖 Привет! Я — бот для изучения английского языка!\n\n" +
		"✨ Что я умею:\n" +
		"• 📅 Показывать новое слово\n" +
		"• 🧠 Проводить викторины\n" +
		"• 📚 Помогать запоминать новые слова\n" +
		"• 🔔 Напоминать учиться\n\n" +
		"Нажми кнопку ниже, чтобы начать!"

	keyboard := t.generateMenuKeyboard()

	msg := tgbotapi.NewMessage(message.Chat.ID, welcomeText)
	msg.ReplyMarkup = keyboard

	sendMessage(t.bot, msg)
}

func (t *TelegramAPI) showMainMenu(message *tgbotapi.Message) {
	keyboard := t.generateMenuKeyboard()

	msg := tgbotapi.NewMessage(message.Chat.ID, "🏠 Главное меню:")
	msg.ReplyMarkup = keyboard

	sendMessage(t.bot, msg)
}

func (t *TelegramAPI) generateMenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(ButtonNewWord),
			tgbotapi.NewKeyboardButton(ButtonQuiz),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(ButtonMyWords),
			tgbotapi.NewKeyboardButton(ButtonProgress),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(ButtonHelp),
		),
	)

	keyboard.ResizeKeyboard = true
	keyboard.OneTimeKeyboard = false

	return keyboard
}

func (t *TelegramAPI) handleHelpCommand(message *tgbotapi.Message) {
	helpText := `
📚 Доступные команды:
/start — запустить бота
/help — это сообщение

🎯 Используй кнопки:
• "Слово дня" — новое слово каждый день
• "Викторина" — проверь свои знания
• "Мой прогресс" — сколько слов выучено
• "Помощь" — подсказки и контакты
`

	msg := tgbotapi.NewMessage(message.Chat.ID, helpText)
	sendMessage(t.bot, msg)
}

func (t *TelegramAPI) handleMessage(message *tgbotapi.Message) {
	if message.From == nil {
		log.Printf("Message without sender: %d", message.Chat.ID)
		return
	}
	userID := message.From.ID
	text := message.Text

	switch {
	case text == ButtonNewWord:
		t.word.sendNewWord(message, userID)
	case text == ButtonQuiz:
		t.quiz.sendNewQuiz(message, userID)
	case text == ButtonMyWords:
		t.showMyWordsMenu(message)
	case text == ButtonProgress:
		t.showProgressMenu(message)
	case text == ButtonWordProgress:
		t.word.sendWordStats(message)
	case text == ButtonQuizProgress:
		t.quiz.sendQuizStats(message)
	case text == ButtonLearnedWords:
		t.word.showWords(message, userID, 0, true)
	case text == ButtonNotLearnedWords:
		t.word.showWords(message, userID, 0, false)
	case text == ButtonMainMenu || text == ButtonBack:
		t.showMainMenu(message)
	case text == ButtonHelp:
		t.handleHelpCommand(message)

	default:
		msg := tgbotapi.NewMessage(message.Chat.ID, "Я не понял. Используй кнопки ниже.")
		sendMessage(t.bot, msg)
	}
}

func (t *TelegramAPI) showProgressMenu(message *tgbotapi.Message) {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(ButtonWordProgress),
			tgbotapi.NewKeyboardButton(ButtonQuizProgress),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(ButtonBack),
		),
	)

	keyboard.ResizeKeyboard = true
	keyboard.OneTimeKeyboard = false

	msg := tgbotapi.NewMessage(message.Chat.ID, "Выбери тип статистики:")
	msg.ReplyMarkup = keyboard

	sendMessage(t.bot, msg)
}

func (t *TelegramAPI) showMyWordsMenu(message *tgbotapi.Message) {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(ButtonLearnedWords),
			tgbotapi.NewKeyboardButton(ButtonNotLearnedWords),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(ButtonBack),
		),
	)

	keyboard.ResizeKeyboard = true
	keyboard.OneTimeKeyboard = false

	msg := tgbotapi.NewMessage(message.Chat.ID, "Выбери тип слов:")
	msg.ReplyMarkup = keyboard

	sendMessage(t.bot, msg)
}

func (t *TelegramAPI) handleCallbackQuery(query *tgbotapi.CallbackQuery) {
	callback := tgbotapi.NewCallback(query.ID, "")
	callback.ShowAlert = false
	if _, err := t.bot.Request(callback); err != nil {
		log.Printf("Failed to answer callback: %v", err)
	}

	data := query.Data

	switch {
	case data == "know" || data == "repeat" || data == "new_word":
		t.word.handleWordCallbackQuery(query)

	case strings.HasPrefix(data, "f_") || strings.HasPrefix(data, "t_"):
		t.word.wordHandlePagination(query)

	case strings.HasPrefix(data, "quiz_") || data == "new_quiz":
		t.quiz.handleQuizCallbackQuery(query)

	case data == "main_menu":
		t.showMainMenu(query.Message)

	default:
		log.Printf("Unknown callback data: %s from user %d", data, query.From.ID)
	}
}
