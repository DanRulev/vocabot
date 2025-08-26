package bot

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/DanRulev/vocabot.git/internal/models"
	"github.com/DanRulev/vocabot.git/internal/storage/cache"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type QuizSI interface {
	NewQuiz(ctx context.Context, userID int64) (string, map[string]bool, error)
	AddQuizResult(ctx context.Context, result models.QuizCard) error
	QuizStats(ctx context.Context, userID int64) (string, error)
}

type QuizT struct {
	bot     BotSender
	cache   *cache.Cache
	service QuizSI
}

func NewQuizTAPI(bot BotSender, cache *cache.Cache, service QuizSI) *QuizT {
	return &QuizT{
		bot:     bot,
		cache:   cache,
		service: service,
	}
}

func (t *QuizT) sendNewQuiz(message *tgbotapi.Message, userID int64) {
	ctx, canceled := context.WithTimeout(context.Background(), 10*time.Second)
	defer canceled()

	if message.From == nil {
		log.Printf("Message without sender: %d", message.Chat.ID)
		return
	}

	question, options, err := t.service.NewQuiz(ctx, userID)
	if err != nil {
		log.Printf("failed to get new quiz for chat: %d :%v", message.Chat.ID, err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка при получении викторины. Попробуй позже.")
		sendMessage(t.bot, msg)
		return
	}

	word := models.QuizCard{
		UserID: userID,
		Word:   question,
		Type:   "quiz",
	}

	var buttons [][]tgbotapi.InlineKeyboardButton

	row := make([]tgbotapi.InlineKeyboardButton, 0, 2)
	i := 0

	for answer, isCorrect := range options {
		callbackData := "quiz_wrong"
		if isCorrect {
			word.Translation = answer
			callbackData = "quiz_right"
		}

		button := tgbotapi.NewInlineKeyboardButtonData(answer, callbackData)
		row = append(row, button)
		i++

		if i%2 == 0 {
			buttons = append(buttons, row)
			row = make([]tgbotapi.InlineKeyboardButton, 0, 2)
		}
	}

	if len(row) > 0 {
		buttons = append(buttons, row)
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	msg := tgbotapi.NewMessage(message.Chat.ID, "❓ Как переводится: "+question)
	msg.ParseMode = "markdown"
	msg.ReplyMarkup = &keyboard

	t.cache.SetQuiz(userID, word)

	sendMessage(t.bot, msg)
}

func (t *QuizT) sendQuizStats(message *tgbotapi.Message) {
	userID := message.From.ID
	chatID := message.Chat.ID
	ctx, canceled := context.WithTimeout(context.Background(), 5*time.Second)
	defer canceled()

	stats, err := t.service.QuizStats(ctx, userID)
	if err != nil {
		log.Printf("failed to get quiz stats for user %d: %v", userID, err)
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка получения статистики")
		sendMessage(t.bot, msg)
		return
	}

	msg := tgbotapi.NewMessage(chatID, stats)
	msg.ParseMode = "markdown"

	sendMessage(t.bot, msg)
}

func (t *QuizT) handleQuizCallbackQuery(query *tgbotapi.CallbackQuery) {
	data := query.Data

	switch {
	case data == "new_quiz":
		if query.Message == nil {
			log.Printf("CallbackQuery without message: %v", query.ID)
			return
		}
		t.sendNewQuiz(query.Message, query.From.ID)
	case strings.HasPrefix(data, "quiz_"):
		t.processQuizAnswer(query)
	default:
		log.Printf("Unknown callback data: %s", query.Data)
		msg := tgbotapi.NewMessage(query.Message.Chat.ID, "❌ НЕИЗВЕСТНАЯ КОМАНДА")
		msg.ParseMode = "markdown"
		sendMessage(t.bot, msg)
	}
}

func (t *QuizT) processQuizAnswer(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
	data := query.Data

	quiz, exists := t.cache.GetQuiz(userID)
	if !exists {
		log.Printf("failed to get quiz from cache for user %d", userID)
		msg := tgbotapi.NewMessage(userID, "❌ Не удалось определить викторину.")
		sendMessage(t.bot, msg)
		return
	}

	t.cache.DeleteQuiz(userID)

	quiz.IsCorrect = (data == "quiz_right")

	statusText := "✅ Правильно! " + quiz.Translation
	if !quiz.IsCorrect {
		statusText = "❌ Неправильно. Повтори слово."
	}

	ctx, canceled := context.WithTimeout(context.Background(), 5*time.Second)
	defer canceled()

	err := t.service.AddQuizResult(ctx, quiz)
	if err != nil {
		log.Printf("failed to save quiz result for user %d: %v", userID, err)
	}

	fullText := fmt.Sprintf("%s\n\n%s", query.Message.Text, statusText)
	editMsg := tgbotapi.NewEditMessageText(
		query.Message.Chat.ID,
		query.Message.MessageID,
		fullText,
	)
	editMsg.ParseMode = "markdown"
	var buttons [][]tgbotapi.InlineKeyboardButton
	buttons = append(buttons, []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData("❓ НОВАЯ ВИКТОРИНА", "new_quiz")})

	editMsg.ReplyMarkup = &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: buttons}

	sendMessage(t.bot, editMsg)
}
