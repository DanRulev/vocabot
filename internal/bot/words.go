package bot

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/DanRulev/vocabot.git/internal/models"
	"github.com/DanRulev/vocabot.git/internal/storage/cache"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type WordSI interface {
	RandomWord(ctx context.Context) (string, models.WordCard, error)
	AddWord(ctx context.Context, word models.WordCard) error
	Words(ctx context.Context, userID int64, page int, learned bool) (string, bool, error)
	WordStat(ctx context.Context, userID int64) (string, error)
}

type WordT struct {
	bot     BotSender
	cache   *cache.Cache
	service WordSI
}

func NewWordTAPI(bot BotSender, cache *cache.Cache, service WordSI) *WordT {
	return &WordT{
		bot:     bot,
		cache:   cache,
		service: service,
	}
}

func (t *WordT) sendNewWord(message *tgbotapi.Message, userID int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if message.From == nil {
		log.Printf("Message without sender: %d", message.Chat.ID)
		return
	}

	word, card, err := t.service.RandomWord(ctx)
	if err != nil {
		log.Printf("Failed to get random word for chat %d: %v", message.Chat.ID, err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "Ошибка при получении слова. Попробуй позже.")
		sendMessage(t.bot, msg)
		return
	}

	card.UserID = userID
	t.cache.SetWord(userID, card)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData("✅ Знаю", "know"),
			tgbotapi.NewInlineKeyboardButtonData("❌ Не знаю", "repeat"),
		},
	)

	msg := tgbotapi.NewMessage(message.Chat.ID, word)
	msg.ParseMode = "markdown"
	msg.ReplyMarkup = &keyboard

	sendMessage(t.bot, msg)
}

func (t *WordT) showWords(message *tgbotapi.Message, userID int64, page int, learned bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	text, hasNext, err := t.service.Words(ctx, userID, page, learned) // true = learned
	if err != nil {
		log.Printf("Failed to load words for chat %d: %v", message.Chat.ID, err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка загрузки слов")
		sendMessage(t.bot, msg)
		return
	}

	knowPrefix := "f"
	if learned {
		knowPrefix = "t"
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = "markdown"
	keyboard := t.wordPaginationKeyboard(knowPrefix, page, hasNext)
	if keyboard != nil {
		msg.ReplyMarkup = keyboard
	}
	sendMessage(t.bot, msg)
}

func (t *WordT) sendWordStats(message *tgbotapi.Message) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stats, err := t.service.WordStat(ctx, message.From.ID)
	if err != nil {
		log.Printf("Failed to get stats for chat %d: %v", message.Chat.ID, err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка")
		sendMessage(t.bot, msg)
		return
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, stats)
	msg.ParseMode = "markdown"
	sendMessage(t.bot, msg)
}

func (t *WordT) handleWordCallbackQuery(query *tgbotapi.CallbackQuery) {
	data := query.Data

	switch data {
	case "know", "repeat":
		t.handleWordResponse(query)
	case "new_word":
		if query.Message == nil {
			log.Printf("CallbackQuery without message: %v", query.ID)
			return
		}
		t.sendNewWord(query.Message, query.From.ID)
	default:
		log.Printf("Unknown callback data: %s", query.Data)
		msg := tgbotapi.NewMessage(query.Message.Chat.ID, "❌ НЕИЗВЕСТНАЯ КОМАНДА")
		msg.ParseMode = "markdown"
		sendMessage(t.bot, msg)
	}
}

func (t *WordT) handleWordResponse(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
	data := query.Data

	word, exists := t.cache.GetWord(userID)
	if !exists {
		msg := tgbotapi.NewMessage(userID, "Не удалось определить слово.")
		sendMessage(t.bot, msg)
		return
	}

	t.cache.DeleteWord(userID)

	var statusText string
	switch data {
	case "know":
		word.Known = true
		statusText = "✅ Отлично! Слово отмечено как выученное."
	case "repeat":
		word.Known = false
		statusText = "❌ Запомнили. Повтори позже."
	default:
		return
	}

	word.UserID = userID

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := t.service.AddWord(ctx, word); err != nil {
		log.Printf("Failed to save word for user %d: %v", userID, err)
	}

	fullText := fmt.Sprintf("%s\n\n%s", query.Message.Text, statusText)
	editMsg := tgbotapi.NewEditMessageText(
		query.Message.Chat.ID,
		query.Message.MessageID,
		fullText,
	)
	editMsg.ParseMode = "markdown"

	var buttons [][]tgbotapi.InlineKeyboardButton
	buttons = append(buttons, []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData("❓ НОВОЕ СЛОВО", "new_word")})

	editMsg.ReplyMarkup = &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: buttons}

	sendMessage(t.bot, editMsg)
}

func (t *WordT) wordHandlePagination(query *tgbotapi.CallbackQuery) {
	if query.Message == nil {
		log.Printf("CallbackQuery without message from user %d", query.From.ID)
		return
	}
	parts := strings.Split(query.Data, "_")
	if len(parts) < 2 {
		return
	}

	prefix := parts[0]
	if prefix != "f" && prefix != "t" {
		msg := tgbotapi.NewMessage(query.Message.Chat.ID, "❌ Ошибка: неверный формат страницы.")
		sendMessage(t.bot, msg)
		return
	}
	page, err := strconv.Atoi(parts[1])
	if err != nil || page < 0 {
		msg := tgbotapi.NewMessage(query.Message.Chat.ID, "❌ Ошибка: неверный номер страницы.")
		sendMessage(t.bot, msg)
		return
	}

	learned := prefix == "t"

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	text, hasNext, err := t.service.Words(ctx, query.From.ID, page, learned)
	if err != nil {
		msg := tgbotapi.NewMessage(query.Message.Chat.ID, "❌ Ошибка загрузки слов")
		sendMessage(t.bot, msg)
		return
	}
	editMsg := tgbotapi.NewEditMessageText(
		query.Message.Chat.ID,
		query.Message.MessageID,
		text,
	)
	editMsg.ParseMode = "markdown"
	keyboard := t.wordPaginationKeyboard(prefix, page, hasNext)
	if keyboard != nil {
		editMsg.ReplyMarkup = keyboard
	}

	sendMessage(t.bot, editMsg)
}

func (t *WordT) wordPaginationKeyboard(prefix string, page int, hasNxt bool) *tgbotapi.InlineKeyboardMarkup {
	var buttons [][]tgbotapi.InlineKeyboardButton

	row := make([]tgbotapi.InlineKeyboardButton, 0, 2)

	if page > 0 {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("◀️ Назад", fmt.Sprintf("%s_%d", prefix, page-1)))
	}

	if hasNxt {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("Далее ▶️", fmt.Sprintf("%s_%d", prefix, page+1)))
	}

	if len(row) > 0 {
		buttons = append(buttons, row)
	}

	buttons = append(buttons, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("❓ НОВОЕ СЛОВО", "new_word"),
		tgbotapi.NewInlineKeyboardButtonData("🏠 Главное меню", "main_menu"),
	})

	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: buttons}
}
