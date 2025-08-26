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

func newQuizTMock(t *testing.T, ctrl *gomock.Controller, setupMock func(*mock_bot.MockServiceI, *mock_bot.MockBot)) *QuizT {
	mockService := mock_bot.NewMockServiceI(ctrl)
	cache := cache.NewCache()
	mockBot := &mock_bot.MockBot{}

	if setupMock != nil {
		setupMock(mockService, mockBot)
	}

	return NewQuizTAPI(mockBot, cache, mockService)
}

func TestQuizT_sendNewQuiz(t *testing.T) {
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
	}{
		{
			name: "success: sends quiz with options",
			args: args{
				message: &tgbotapi.Message{
					Chat: &tgbotapi.Chat{ID: 123},
					From: &tgbotapi.User{ID: 456},
				},
				userID: 456,
			},
			f: func(ms *mock_bot.MockServiceI, mb *mock_bot.MockBot) {
				options := map[string]bool{
					"Привет":     true,
					"Пока":       false,
					"Здрасьте":   false,
					"Досвидания": false,
				}
				ms.EXPECT().NewQuiz(gomock.Any(), int64(456)).Return("hello", options, nil)
			},
			assertFunc: func(t *testing.T, mb *mock_bot.MockBot) {
				require.Equal(t, 1, len(mb.SentMessages))
				msg, ok := mb.SentMessages[0].(tgbotapi.MessageConfig)
				require.True(t, ok)
				assert.Equal(t, "❓ Как переводится: hello", msg.Text)
				assert.Equal(t, "markdown", msg.ParseMode)
				assert.NotNil(t, msg.ReplyMarkup)
			},
		},
		{
			name: "error: NewQuiz fails",
			args: args{
				message: &tgbotapi.Message{
					Chat: &tgbotapi.Chat{ID: 123},
					From: &tgbotapi.User{ID: 456},
				},
				userID: 456,
			},
			f: func(ms *mock_bot.MockServiceI, mb *mock_bot.MockBot) {
				ms.EXPECT().NewQuiz(gomock.Any(), int64(456)).Return("", nil, assert.AnError)
			},
			assertFunc: func(t *testing.T, mb *mock_bot.MockBot) {
				require.Equal(t, 1, len(mb.SentMessages))
				msg := mb.SentMessages[0].(tgbotapi.MessageConfig)
				assert.Equal(t, "❌ Ошибка при получении викторины. Попробуй позже.", msg.Text)
			},
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
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			quizT := newQuizTMock(t, ctrl, tt.f)
			mb, _ := quizT.bot.(*mock_bot.MockBot)

			mock_bot.ClearSentMessages(mb)
			quizT.sendNewQuiz(tt.args.message, tt.args.userID)

			if tt.assertFunc != nil {
				tt.assertFunc(t, mb)
			}
		})
	}
}
func TestQuizT_processQuizAnswer(t *testing.T) {
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
			name: "correct answer: quiz_right",
			args: args{
				query: &tgbotapi.CallbackQuery{
					From: &tgbotapi.User{ID: 456},
					Message: &tgbotapi.Message{
						Chat:      &tgbotapi.Chat{ID: 123},
						MessageID: 100,
						Text:      "❓ Как переводится: hello",
					},
					Data: "quiz_right",
				},
			},
			f: func(ms *mock_bot.MockServiceI, mb *mock_bot.MockBot) {
				ms.EXPECT().AddQuizResult(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, card models.QuizCard) error {
						assert.True(t, card.IsCorrect)
						assert.Equal(t, int64(456), card.UserID)
						return nil
					},
				)
			},
			assertFunc: func(t *testing.T, mb *mock_bot.MockBot) {
				require.Equal(t, 1, len(mb.SentMessages))
				editMsg := mb.SentMessages[0].(tgbotapi.EditMessageTextConfig)
				assert.Contains(t, editMsg.Text, "✅ Правильно! ")
				assert.NotNil(t, editMsg.ReplyMarkup)
				kb := editMsg.ReplyMarkup
				assert.Equal(t, "❓ НОВАЯ ВИКТОРИНА", kb.InlineKeyboard[0][0].Text)
			},
		},
		{
			name: "wrong answer: quiz_wrong",
			args: args{
				query: &tgbotapi.CallbackQuery{
					From: &tgbotapi.User{ID: 456},
					Message: &tgbotapi.Message{
						Chat:      &tgbotapi.Chat{ID: 123},
						MessageID: 100,
						Text:      "❓ Как переводится: hello",
					},
					Data: "quiz_wrong",
				},
			},
			f: func(ms *mock_bot.MockServiceI, mb *mock_bot.MockBot) {
				ms.EXPECT().AddQuizResult(gomock.Any(), gomock.Any()).Return(nil)
			},
			assertFunc: func(t *testing.T, mb *mock_bot.MockBot) {
				editMsg := mb.SentMessages[0].(tgbotapi.EditMessageTextConfig)
				assert.Contains(t, editMsg.Text, "❌ Неправильно. Повтори слово.")
			},
		},
		{
			name: "no quiz in cache",
			args: args{
				query: &tgbotapi.CallbackQuery{
					From: &tgbotapi.User{ID: 456},
					Message: &tgbotapi.Message{
						Chat: &tgbotapi.Chat{ID: 123},
					},
					Data: "quiz_right",
				},
			},
			assertFunc: func(t *testing.T, mb *mock_bot.MockBot) {
				require.Equal(t, 1, len(mb.SentMessages))
				msg := mb.SentMessages[0].(tgbotapi.MessageConfig)
				assert.Equal(t, "❌ Не удалось определить викторину.", msg.Text)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			quizT := newQuizTMock(t, ctrl, tt.f)
			mb, _ := quizT.bot.(*mock_bot.MockBot)

			if tt.name != "no quiz in cache" {
				quizT.cache.SetQuiz(456, models.QuizCard{
					UserID:      456,
					Word:        "hello",
					Translation: "Привет",
					Type:        "quiz",
				})
			}

			mock_bot.ClearSentMessages(mb)
			quizT.processQuizAnswer(tt.args.query)

			if tt.assertFunc != nil {
				tt.assertFunc(t, mb)
			}
		})
	}
}

func TestQuizT_sendQuizStats(t *testing.T) {
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
				ms.EXPECT().QuizStats(gomock.Any(), int64(456)).Return("📊 Викторин пройдено: 10\n✅ Правильно: 8", nil)
			},
			assertFunc: func(t *testing.T, mb *mock_bot.MockBot) {
				require.Equal(t, 1, len(mb.SentMessages))
				msg := mb.SentMessages[0].(tgbotapi.MessageConfig)
				assert.Equal(t, "📊 Викторин пройдено: 10\n✅ Правильно: 8", msg.Text)
				assert.Equal(t, "markdown", msg.ParseMode)
			},
		},
		{
			name: "error: failed to get stats",
			f: func(ms *mock_bot.MockServiceI, mb *mock_bot.MockBot) {
				ms.EXPECT().QuizStats(gomock.Any(), int64(456)).Return("", assert.AnError)
			},
			assertFunc: func(t *testing.T, mb *mock_bot.MockBot) {
				msg := mb.SentMessages[0].(tgbotapi.MessageConfig)
				assert.Equal(t, "❌ Ошибка получения статистики", msg.Text)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			quizT := newQuizTMock(t, ctrl, tt.f)
			mb, _ := quizT.bot.(*mock_bot.MockBot)

			mock_bot.ClearSentMessages(mb)
			quizT.sendQuizStats(message)

			if tt.assertFunc != nil {
				tt.assertFunc(t, mb)
			}
		})
	}
}

func TestQuizT_handleQuizCallbackQuery(t *testing.T) {
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
			name: "new_quiz: triggers sendNewQuiz",
			args: args{
				query: &tgbotapi.CallbackQuery{
					From: &tgbotapi.User{ID: 456},
					Message: &tgbotapi.Message{
						Chat: &tgbotapi.Chat{ID: 123},
						From: &tgbotapi.User{ID: 456}, // ✅ Добавь это!
					},
					Data: "new_quiz",
				},
			},
			f: func(ms *mock_bot.MockServiceI, mb *mock_bot.MockBot) {
				ms.EXPECT().NewQuiz(gomock.Any(), int64(456)).Return("hello", map[string]bool{"Привет": true}, nil)
			},
			assertFunc: func(t *testing.T, mb *mock_bot.MockBot) {
				require.Equal(t, 1, len(mb.SentMessages))
				msg := mb.SentMessages[0].(tgbotapi.MessageConfig)
				assert.Contains(t, msg.Text, "Как переводится: hello")
			},
		},
		{
			name: "quiz_right: processes answer",
			args: args{
				query: &tgbotapi.CallbackQuery{
					From: &tgbotapi.User{ID: 456},
					Message: &tgbotapi.Message{
						Chat: &tgbotapi.Chat{ID: 123},
						Text: "❓ Как переводится: hello",
					},
					Data: "quiz_right",
				},
			},
			f: func(ms *mock_bot.MockServiceI, mb *mock_bot.MockBot) {
				ms.EXPECT().AddQuizResult(gomock.Any(), gomock.Any()).Return(nil)
			},
			assertFunc: func(t *testing.T, mb *mock_bot.MockBot) {
				require.Equal(t, 1, len(mb.SentMessages))
				editMsg := mb.SentMessages[0].(tgbotapi.EditMessageTextConfig)
				assert.Contains(t, editMsg.Text, "✅ Правильно!")
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
				msg := mb.SentMessages[0].(tgbotapi.MessageConfig)
				assert.Equal(t, "❌ НЕИЗВЕСТНАЯ КОМАНДА", msg.Text)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			quizT := newQuizTMock(t, ctrl, tt.f)
			mb, _ := quizT.bot.(*mock_bot.MockBot)

			if tt.args.query.Data == "quiz_right" {
				quizT.cache.SetQuiz(456, models.QuizCard{
					UserID:      456,
					Word:        "hello",
					Translation: "Привет",
				})
			}

			mock_bot.ClearSentMessages(mb)
			quizT.handleQuizCallbackQuery(tt.args.query)

			if tt.assertFunc != nil {
				tt.assertFunc(t, mb)
			}
		})
	}
}
