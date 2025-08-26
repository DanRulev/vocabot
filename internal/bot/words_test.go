package bot

import (
	"context"
	"testing"

	mock_bot "github.com/DanRulev/vocabot.git/internal/bot/mock"
	"github.com/DanRulev/vocabot.git/internal/models"
	"github.com/DanRulev/vocabot.git/internal/storage/cache"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newWordTMock(t *testing.T, ctrl *gomock.Controller, setupMock func(*mock_bot.MockServiceI, *mock_bot.MockBot)) *WordT {
	mockService := mock_bot.NewMockServiceI(ctrl)
	cache := cache.NewCache()
	mockBot := &mock_bot.MockBot{}

	if setupMock != nil {
		setupMock(mockService, mockBot)
	}

	return NewWordTAPI(mockBot, cache, mockService)
}

func TestWordT_sendNewWord(t *testing.T) {
	t.Parallel()

	type args struct {
		message *tgbotapi.Message
		userID  int64
	}
	tests := []struct {
		name       string
		args       args
		f          func(*mock_bot.MockServiceI, *mock_bot.MockBot)
		assertFunc func(*testing.T, *mock_bot.MockBot)
		wantErr    bool
	}{
		{
			name: "success: sends word and keyboard",
			args: args{
				message: &tgbotapi.Message{
					Chat: &tgbotapi.Chat{ID: 123},
					From: &tgbotapi.User{ID: 456},
				},
				userID: 456,
			},
			f: func(ms *mock_bot.MockServiceI, mb *mock_bot.MockBot) {
				ms.EXPECT().RandomWord(gomock.Any()).Return(
					"**hello**\n*привет*",
					models.WordCard{WordText: "hello", Translation: "привет"},
					nil,
				)
			},
			assertFunc: func(t *testing.T, mb *mock_bot.MockBot) {
				require.Equal(t, 1, len(mb.SentMessages))
				msg, ok := mb.SentMessages[0].(tgbotapi.MessageConfig)
				require.True(t, ok)
				assert.Equal(t, "**hello**\n*привет*", msg.Text)
				assert.Equal(t, "markdown", msg.ParseMode)
				assert.NotNil(t, msg.ReplyMarkup)
				keyboard := msg.ReplyMarkup
				kb, ok := keyboard.(*tgbotapi.InlineKeyboardMarkup)
				require.True(t, ok)
				assert.Equal(t, 1, len(kb.InlineKeyboard))
				assert.Equal(t, "✅ Знаю", kb.InlineKeyboard[0][0].Text)
				assert.Equal(t, "❌ Не знаю", kb.InlineKeyboard[0][1].Text)
			},
			wantErr: false,
		},
		{
			name: "error: RandomWord fails",
			args: args{
				message: &tgbotapi.Message{
					Chat: &tgbotapi.Chat{ID: 123},
					From: &tgbotapi.User{ID: 456},
				},
				userID: 456,
			},
			f: func(ms *mock_bot.MockServiceI, mb *mock_bot.MockBot) {
				ms.EXPECT().RandomWord(gomock.Any()).Return("", models.WordCard{}, assert.AnError)
			},
			assertFunc: func(t *testing.T, mb *mock_bot.MockBot) {
				require.Equal(t, 1, len(mb.SentMessages))
				msg, ok := mb.SentMessages[0].(tgbotapi.MessageConfig)
				require.True(t, ok)
				assert.Equal(t, "Ошибка при получении слова. Попробуй позже.", msg.Text)
			},
			wantErr: false,
		},
		{
			name: "nil From in message",
			args: args{
				message: &tgbotapi.Message{
					Chat: &tgbotapi.Chat{ID: 123},
					From: nil,
				},
				userID: 456,
			},
			assertFunc: func(t *testing.T, mb *mock_bot.MockBot) {
				assert.Empty(t, mb.SentMessages)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			wordT := newWordTMock(t, ctrl, tt.f)
			mb, _ := wordT.bot.(*mock_bot.MockBot)

			mock_bot.ClearSentMessages(mb)
			wordT.sendNewWord(tt.args.message, tt.args.userID)

			if tt.assertFunc != nil {
				tt.assertFunc(t, mb)
			}
		})
	}
}

func TestWordT_handleWordResponse(t *testing.T) {
	t.Parallel()

	type args struct {
		query *tgbotapi.CallbackQuery
	}

	tests := []struct {
		name       string
		args       args
		f          func(*mock_bot.MockServiceI, *mock_bot.MockBot)
		assertFunc func(*testing.T, *mock_bot.MockBot)
	}{
		{
			name: "know: marks word as known",
			args: args{
				query: &tgbotapi.CallbackQuery{
					From: &tgbotapi.User{ID: 456},
					Message: &tgbotapi.Message{
						Chat:      &tgbotapi.Chat{ID: 123},
						MessageID: 100,
						Text:      "**hello**",
					},
					Data: "know",
				},
			},
			f: func(ms *mock_bot.MockServiceI, mb *mock_bot.MockBot) {
				ms.EXPECT().AddWord(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, word models.WordCard) error {
						assert.True(t, word.Known)
						assert.Equal(t, int64(456), word.UserID)
						return nil
					},
				)
			},
			assertFunc: func(t *testing.T, mb *mock_bot.MockBot) {
				require.Equal(t, 1, len(mb.SentMessages))
				editMsg, ok := mb.SentMessages[0].(tgbotapi.EditMessageTextConfig)
				require.True(t, ok)
				assert.Contains(t, editMsg.Text, "✅ Отлично! Слово отмечено как выученное.")
				assert.NotNil(t, editMsg.ReplyMarkup)
				kb := editMsg.ReplyMarkup
				assert.Equal(t, "❓ НОВОЕ СЛОВО", kb.InlineKeyboard[0][0].Text)
			},
		},
		{
			name: "repeat: marks as not known",
			args: args{
				query: &tgbotapi.CallbackQuery{
					From: &tgbotapi.User{ID: 456},
					Message: &tgbotapi.Message{
						Chat:      &tgbotapi.Chat{ID: 123},
						MessageID: 100,
						Text:      "**hello**",
					},
					Data: "repeat",
				},
			},
			f: func(ms *mock_bot.MockServiceI, mb *mock_bot.MockBot) {
				ms.EXPECT().AddWord(gomock.Any(), gomock.Any()).Return(nil)
			},
			assertFunc: func(t *testing.T, mb *mock_bot.MockBot) {
				editMsg := mb.SentMessages[0].(tgbotapi.EditMessageTextConfig)
				assert.Contains(t, editMsg.Text, "❌ Запомнили. Повтори позже.")
			},
		},
		{
			name: "no word in cache",
			args: args{
				query: &tgbotapi.CallbackQuery{
					From: &tgbotapi.User{ID: 456},
					Message: &tgbotapi.Message{
						Chat: &tgbotapi.Chat{ID: 123},
					},
					Data: "know",
				},
			},
			assertFunc: func(t *testing.T, mb *mock_bot.MockBot) {
				require.Equal(t, 1, len(mb.SentMessages))
				msg := mb.SentMessages[0].(tgbotapi.MessageConfig)
				assert.Equal(t, "Не удалось определить слово.", msg.Text)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			wordT := newWordTMock(t, ctrl, tt.f)
			mb, _ := wordT.bot.(*mock_bot.MockBot)

			if tt.name != "no word in cache" {
				wordT.cache.SetWord(456, models.WordCard{WordText: "hello", Translation: "привет"})
			}

			mock_bot.ClearSentMessages(mb)
			wordT.handleWordResponse(tt.args.query)

			if tt.assertFunc != nil {
				tt.assertFunc(t, mb)
			}
		})
	}
}

func TestWordT_showWords(t *testing.T) {
	t.Parallel()

	type args struct {
		message *tgbotapi.Message
		userID  int64
		page    int
		learned bool
	}

	tests := []struct {
		name       string
		args       args
		f          func(*mock_bot.MockServiceI, *mock_bot.MockBot)
		assertFunc func(*testing.T, *mock_bot.MockBot)
	}{
		{
			name: "learned words: shows list and pagination",
			args: args{
				message: &tgbotapi.Message{
					Chat: &tgbotapi.Chat{ID: 123},
					From: &tgbotapi.User{ID: 456},
				},
				userID:  456,
				page:    0,
				learned: true,
			},
			f: func(ms *mock_bot.MockServiceI, mb *mock_bot.MockBot) {
				ms.EXPECT().Words(gomock.Any(), int64(456), 0, true).Return("✅ Выученные: 5 слов", true, nil)
			},
			assertFunc: func(t *testing.T, mb *mock_bot.MockBot) {
				msg := mb.SentMessages[0].(tgbotapi.MessageConfig)
				assert.Equal(t, "✅ Выученные: 5 слов", msg.Text)
				assert.NotNil(t, msg.ReplyMarkup)
				kb := msg.ReplyMarkup.(*tgbotapi.InlineKeyboardMarkup)
				assert.Equal(t, "Далее ▶️", kb.InlineKeyboard[0][0].Text)
				assert.Equal(t, "❓ НОВОЕ СЛОВО", kb.InlineKeyboard[1][0].Text)
			},
		},
		{
			name: "error: failed to load words",
			args: args{
				message: &tgbotapi.Message{
					Chat: &tgbotapi.Chat{ID: 123},
					From: &tgbotapi.User{ID: 456},
				},
				userID:  456,
				page:    0,
				learned: false,
			},
			f: func(ms *mock_bot.MockServiceI, mb *mock_bot.MockBot) {
				ms.EXPECT().Words(gomock.Any(), int64(456), 0, false).Return("", false, assert.AnError)
			},
			assertFunc: func(t *testing.T, mb *mock_bot.MockBot) {
				msg := mb.SentMessages[0].(tgbotapi.MessageConfig)
				assert.Equal(t, "❌ Ошибка загрузки слов", msg.Text)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			wordT := newWordTMock(t, ctrl, tt.f)
			mb, _ := wordT.bot.(*mock_bot.MockBot)

			mock_bot.ClearSentMessages(mb)
			wordT.showWords(tt.args.message, tt.args.userID, tt.args.page, tt.args.learned)

			if tt.assertFunc != nil {
				tt.assertFunc(t, mb)
			}
		})
	}
}

func TestWordT_sendWordStats(t *testing.T) {
	t.Parallel()

	message := &tgbotapi.Message{
		Chat: &tgbotapi.Chat{ID: 123},
		From: &tgbotapi.User{ID: 456},
	}

	tests := []struct {
		name       string
		f          func(*mock_bot.MockServiceI, *mock_bot.MockBot)
		assertFunc func(*testing.T, *mock_bot.MockBot)
	}{
		{
			name: "success: sends stats",
			f: func(ms *mock_bot.MockServiceI, mb *mock_bot.MockBot) {
				ms.EXPECT().WordStat(gomock.Any(), int64(456)).Return("📊 Слов выучено: 10", nil)
			},
			assertFunc: func(t *testing.T, mb *mock_bot.MockBot) {
				msg := mb.SentMessages[0].(tgbotapi.MessageConfig)
				assert.Equal(t, "📊 Слов выучено: 10", msg.Text)
				assert.Equal(t, "markdown", msg.ParseMode)
			},
		},
		{
			name: "error: failed to get stats",
			f: func(ms *mock_bot.MockServiceI, mb *mock_bot.MockBot) {
				ms.EXPECT().WordStat(gomock.Any(), int64(456)).Return("", assert.AnError)
			},
			assertFunc: func(t *testing.T, mb *mock_bot.MockBot) {
				msg := mb.SentMessages[0].(tgbotapi.MessageConfig)
				assert.Equal(t, "❌ Ошибка", msg.Text)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			wordT := newWordTMock(t, ctrl, tt.f)
			mb, _ := wordT.bot.(*mock_bot.MockBot)

			mock_bot.ClearSentMessages(mb)
			wordT.sendWordStats(message)

			if tt.assertFunc != nil {
				tt.assertFunc(t, mb)
			}
		})
	}
}

func TestWordT_handleWordCallbackQuery(t *testing.T) {
	t.Parallel()

	type args struct {
		query *tgbotapi.CallbackQuery
	}

	tests := []struct {
		name       string
		args       args
		f          func(*mock_bot.MockServiceI, *mock_bot.MockBot)
		assertFunc func(*testing.T, *mock_bot.MockBot)
	}{
		{
			name: "know: calls handleWordResponse",
			args: args{
				query: &tgbotapi.CallbackQuery{
					From: &tgbotapi.User{ID: 456},
					Message: &tgbotapi.Message{
						Chat:      &tgbotapi.Chat{ID: 123},
						MessageID: 100,
						Text:      "hello",
					},
					Data: "know",
				},
			},
			f: func(ms *mock_bot.MockServiceI, mb *mock_bot.MockBot) {
				ms.EXPECT().AddWord(gomock.Any(), gomock.Any()).Return(nil)
			},
			assertFunc: func(t *testing.T, mb *mock_bot.MockBot) {
				require.Equal(t, 1, len(mb.SentMessages))
				editMsg, ok := mb.SentMessages[0].(tgbotapi.EditMessageTextConfig)
				require.True(t, ok)
				assert.Contains(t, editMsg.Text, "✅ Отлично! Слово отмечено как выученное.")
			},
		},
		{
			name: "repeat: calls handleWordResponse",
			args: args{
				query: &tgbotapi.CallbackQuery{
					From: &tgbotapi.User{ID: 456},
					Message: &tgbotapi.Message{
						Chat:      &tgbotapi.Chat{ID: 123},
						MessageID: 100,
						Text:      "hello",
					},
					Data: "repeat",
				},
			},
			f: func(ms *mock_bot.MockServiceI, mb *mock_bot.MockBot) {
				ms.EXPECT().AddWord(gomock.Any(), gomock.Any()).Return(nil)
			},
			assertFunc: func(t *testing.T, mb *mock_bot.MockBot) {
				editMsg := mb.SentMessages[0].(tgbotapi.EditMessageTextConfig)
				assert.Contains(t, editMsg.Text, "❌ Запомнили. Повтори позже.")
			},
		},
		{
			name: "new_word: calls sendNewWord",
			args: args{
				query: &tgbotapi.CallbackQuery{
					From: &tgbotapi.User{ID: 456},
					Message: &tgbotapi.Message{
						Chat: &tgbotapi.Chat{ID: 123},
						From: &tgbotapi.User{ID: 456},
					},
					Data: "new_word",
				},
			},
			f: func(ms *mock_bot.MockServiceI, mb *mock_bot.MockBot) {
				ms.EXPECT().RandomWord(gomock.Any()).Return(
					"**hello**\n*привет*",
					models.WordCard{WordText: "hello", Translation: "привет"},
					nil,
				)
			},
			assertFunc: func(t *testing.T, mb *mock_bot.MockBot) {
				require.Equal(t, 1, len(mb.SentMessages))
				msg, ok := mb.SentMessages[0].(tgbotapi.MessageConfig)
				require.True(t, ok)
				assert.Equal(t, "**hello**\n*привет*", msg.Text)
			},
		},
		{
			name: "new_word: query.Message is nil",
			args: args{
				query: &tgbotapi.CallbackQuery{
					From: &tgbotapi.User{ID: 456},
					Data: "new_word",
				},
			},
			assertFunc: func(t *testing.T, mb *mock_bot.MockBot) {
				assert.Empty(t, mb.SentMessages)
			},
		},
		{
			name: "unknown command",
			args: args{
				query: &tgbotapi.CallbackQuery{
					From: &tgbotapi.User{ID: 456},
					Message: &tgbotapi.Message{
						Chat: &tgbotapi.Chat{ID: 123},
					},
					Data: "unknown_data",
				},
			},
			assertFunc: func(t *testing.T, mb *mock_bot.MockBot) {
				require.Equal(t, 1, len(mb.SentMessages))
				msg := mb.SentMessages[0].(tgbotapi.MessageConfig)
				assert.Equal(t, "❌ НЕИЗВЕСТНАЯ КОМАНДА", msg.Text)
				assert.Equal(t, "markdown", msg.ParseMode)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			wordT := newWordTMock(t, ctrl, tt.f)
			mb, _ := wordT.bot.(*mock_bot.MockBot)

			if tt.name == "know: calls handleWordResponse" || tt.name == "repeat: calls handleWordResponse" {
				wordT.cache.SetWord(456, models.WordCard{WordText: "hello", Translation: "привет"})
			}

			mock_bot.ClearSentMessages(mb)
			wordT.handleWordCallbackQuery(tt.args.query)

			if tt.assertFunc != nil {
				tt.assertFunc(t, mb)
			}
		})
	}
}

func TestWordT_wordHandlePagination(t *testing.T) {
	t.Parallel()

	type args struct {
		query *tgbotapi.CallbackQuery
	}

	tests := []struct {
		name       string
		args       args
		f          func(*mock_bot.MockServiceI, *mock_bot.MockBot)
		assertFunc func(*testing.T, *mock_bot.MockBot)
	}{
		{
			name: "pagination: next page (f_1)",
			args: args{
				query: &tgbotapi.CallbackQuery{
					From: &tgbotapi.User{ID: 456},
					Message: &tgbotapi.Message{
						Chat:      &tgbotapi.Chat{ID: 123},
						MessageID: 100,
					},
					Data: "f_1",
				},
			},
			f: func(ms *mock_bot.MockServiceI, mb *mock_bot.MockBot) {
				ms.EXPECT().Words(gomock.Any(), int64(456), 1, false).Return("Слово 1\nСлово 2", false, nil)
			},
			assertFunc: func(t *testing.T, mb *mock_bot.MockBot) {
				require.Equal(t, 1, len(mb.SentMessages))
				editMsg := mb.SentMessages[0].(tgbotapi.EditMessageTextConfig)
				assert.Equal(t, "Слово 1\nСлово 2", editMsg.Text)
				assert.NotNil(t, editMsg.ReplyMarkup)
				kb := editMsg.ReplyMarkup
				assert.Equal(t, "❓ НОВОЕ СЛОВО", kb.InlineKeyboard[1][0].Text)
			},
		},
		{
			name: "invalid page format",
			args: args{
				query: &tgbotapi.CallbackQuery{
					From: &tgbotapi.User{ID: 456},
					Message: &tgbotapi.Message{
						Chat: &tgbotapi.Chat{ID: 123},
					},
					Data: "x_abc",
				},
			},
			assertFunc: func(t *testing.T, mb *mock_bot.MockBot) {
				msg := mb.SentMessages[0].(tgbotapi.MessageConfig)
				assert.Equal(t, "❌ Ошибка: неверный формат страницы.", msg.Text)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			wordT := newWordTMock(t, ctrl, tt.f)
			mb, _ := wordT.bot.(*mock_bot.MockBot)

			mock_bot.ClearSentMessages(mb)
			wordT.wordHandlePagination(tt.args.query)

			if tt.assertFunc != nil {
				tt.assertFunc(t, mb)
			}
		})
	}
}
