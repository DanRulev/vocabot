package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DanRulev/vocabot.git/internal/models"
	mock_service "github.com/DanRulev/vocabot.git/internal/service/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newWordServiceMock(t *testing.T, ctrl *gomock.Controller, setupMock func(*mock_service.MockRepositoryI, *mock_service.MockAPII)) *WordS {
	api := mock_service.NewMockAPII(ctrl)
	repo := mock_service.NewMockRepositoryI(ctrl)
	if setupMock != nil {
		setupMock(repo, api)
	}

	log := zap.NewNop()

	return &WordS{
		myMemory:       api,
		pythonAnyWhere: api,
		vercel:         api,
		repo:           repo,
		log:            log,
	}
}

func TestWordS_RandomWord(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name       string
		args       args
		f          func(*mock_service.MockRepositoryI, *mock_service.MockAPII)
		assertFunc func(t *testing.T, result string, card models.WordCard, err error)
		wantErr    bool
	}{
		{
			name: "success",
			args: args{ctx: context.Background()},
			f: func(mri *mock_service.MockRepositoryI, ma *mock_service.MockAPII) {
				ma.EXPECT().RandomWord(gomock.Any()).Return("hello", nil)
				ma.EXPECT().TranslateEnToRu(gomock.Any(), "hello").Return(models.MyMemoryTranslationResult{
					Text:     "привет",
					Match:    0.9,
					Reliable: true,
				}, nil)
				ma.EXPECT().DictionaryData(gomock.Any(), "hello").Return(models.TranslationResponse{
					SourceText:      "hello",
					DestinationText: "привет",
					Pronunciation: struct {
						SourceTextPhonetic   string `json:"source-text-phonetic"`
						SourceTextAudio      string `json:"source-text-audio"`
						DestinationTextAudio string `json:"destination-text-audio"`
					}{
						SourceTextPhonetic: "[həˈləʊ]",
					},
					Definitions: []struct {
						PartOfSpeech  string              `json:"part-of-speech"`
						Definition    string              `json:"definition"`
						Example       string              `json:"example"`
						OtherExamples []string            `json:"other-examples"`
						Synonyms      map[string][]string `json:"synonyms"`
					}{
						{
							PartOfSpeech: "noun",
							Definition:   "a greeting when meeting someone",
							Example:      "Hello, how are you?",
						},
						{
							PartOfSpeech: "interjection",
							Definition:   "used to express surprise",
							Example:      "Hello! What are you doing here?",
						},
					},
				}, nil)
			},
			wantErr: false,
			assertFunc: func(t *testing.T, result string, card models.WordCard, err error) {
				assert.Contains(t, result, "**hello**")
				assert.Contains(t, result, "привет")
				assert.Contains(t, result, "`[həˈləʊ]`")
				assert.Contains(t, result, "🔖 *noun*")
				assert.Contains(t, result, "a greeting")
				assert.Contains(t, result, "Hello, how are you?")
				assert.Contains(t, result, "interjection")

				assert.Equal(t, "hello", card.WordText)
				assert.Equal(t, "привет", card.Translation)
			},
		},
		{
			name: "success: word without definitions",
			f: func(mri *mock_service.MockRepositoryI, ma *mock_service.MockAPII) {
				ma.EXPECT().RandomWord(gomock.Any()).Return("xyz", nil)
				ma.EXPECT().TranslateEnToRu(gomock.Any(), "xyz").Return(models.MyMemoryTranslationResult{
					Text: "абв",
				}, nil)

				ma.EXPECT().DictionaryData(gomock.Any(), "xyz").Return(models.TranslationResponse{
					SourceText:      "xyz",
					DestinationText: "абв",
					Definitions:     nil,
				}, nil)
			},
			wantErr: false,
			assertFunc: func(t *testing.T, result string, card models.WordCard, err error) {
				assert.Contains(t, result, "**xyz**")
				assert.Contains(t, result, "абв")                      // ✅ без *
				assert.Contains(t, result, "⚠️ Нет словарных данных.") // ✅ совпадает с formatTranslation
				assert.NotContains(t, result, "🔖")

				assert.Equal(t, "xyz", card.WordText)
				assert.Equal(t, "абв", card.Translation)
			},
		},
		{
			name: "success: retry then succeed",
			f: func(mri *mock_service.MockRepositoryI, ma *mock_service.MockAPII) {
				ma.EXPECT().RandomWord(gomock.Any()).Return("fail", nil)
				ma.EXPECT().TranslateEnToRu(gomock.Any(), "fail").Return(models.MyMemoryTranslationResult{}, errors.New("temp error"))

				ma.EXPECT().RandomWord(gomock.Any()).Return("empty", nil)
				ma.EXPECT().TranslateEnToRu(gomock.Any(), "empty").Return(models.MyMemoryTranslationResult{Text: ""}, nil)

				ma.EXPECT().RandomWord(gomock.Any()).Return("success", nil)
				ma.EXPECT().TranslateEnToRu(gomock.Any(), "success").Return(models.MyMemoryTranslationResult{
					Text: "успех",
				}, nil)

				ma.EXPECT().DictionaryData(gomock.Any(), "success").Return(models.TranslationResponse{
					SourceText:      "success",
					DestinationText: "успех",
				}, nil)
			},
			wantErr: false,
			assertFunc: func(t *testing.T, result string, card models.WordCard, err error) {
				assert.Contains(t, result, "успех")
				assert.Equal(t, "success", card.WordText)
				assert.Equal(t, "успех", card.Translation)
			},
		},
		{
			name: "error: RandomWord fails all attempts",
			f: func(mri *mock_service.MockRepositoryI, ma *mock_service.MockAPII) {
				ma.EXPECT().RandomWord(gomock.Any()).Return("", errors.New("service down")).Times(5)
			},
			wantErr: true,
		},
		{
			name: "error: TranslateEnToRu fails all attempts",
			args: args{ctx: context.Background()},
			f: func(mri *mock_service.MockRepositoryI, ma *mock_service.MockAPII) {
				ma.EXPECT().RandomWord(gomock.Any()).Return("hello", nil).Times(5)
				ma.EXPECT().
					TranslateEnToRu(gomock.Any(), gomock.Any()).
					Return(models.MyMemoryTranslationResult{}, errors.New("translation failed")).
					Times(5)
				ma.EXPECT().DictionaryData(gomock.Any(), gomock.Any()).Return(models.TranslationResponse{
					DestinationText: "привет",
				}, nil)
			},
			wantErr: false,
			assertFunc: func(t *testing.T, result string, card models.WordCard, err error) {
				assert.Equal(t, card.Translation, "привет")
			},
		},
		{
			name: "error: empty translation",
			args: args{ctx: context.Background()},
			f: func(mri *mock_service.MockRepositoryI, ma *mock_service.MockAPII) {
				ma.EXPECT().RandomWord(gomock.Any()).Return("bad", nil).Times(5)
				ma.EXPECT().
					TranslateEnToRu(gomock.Any(), gomock.Any()).
					Return(models.MyMemoryTranslationResult{Text: ""}, nil).
					Times(5)
				ma.EXPECT().DictionaryData(gomock.Any(), gomock.Any()).Return(models.TranslationResponse{}, nil)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			wordService := newWordServiceMock(t, ctrl, tt.f)

			got, got1, err := wordService.RandomWord(tt.args.ctx)
			if tt.wantErr {
				require.Error(t, err)
				require.Empty(t, got)
				require.Empty(t, got1)
				return
			}

			require.NoError(t, err)
			if tt.assertFunc != nil {
				tt.assertFunc(t, got, got1, err)
			}
		})
	}
}

func TestEscapeMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "no special characters",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "underscore",
			input:    "hello_world",
			expected: "hello\\_world",
		},
		{
			name:     "asterisk",
			input:    "hello*world",
			expected: "hello\\*world",
		},
		{
			name:     "hash",
			input:    "#hello",
			expected: "\\#hello",
		},
		{
			name:     "exclamation",
			input:    "Hello!",
			expected: "Hello\\!",
		},
		{
			name:     "plus",
			input:    "2+2=4",
			expected: "2+2=4", // не экранируется
		},
		{
			name:     "minus",
			input:    "5-3=2",
			expected: "5-3=2", // не экранируется
		},
		{
			name:     "equals",
			input:    "a=b",
			expected: "a=b", // не экранируется
		},
		{
			name:     "multiple special chars",
			input:    "Hello*world_from#test!value",
			expected: "Hello\\*world\\_from\\#test\\!value",
		},
		{
			name:     "repeated special chars",
			input:    "a__b**c",
			expected: "a\\_\\_b\\*\\*c",
		},
		{
			name:     "only special chars",
			input:    "_*#!",
			expected: "\\_\\*\\#\\!",
		},
		{
			name:     "mixed with normal text",
			input:    "This is a test: 2+2=4, hello_world, #tag, *bold*, !important",
			expected: "This is a test: 2+2=4, hello\\_world, \\#tag, \\*bold\\*, \\!important",
		},
		{
			name:     "no escaping needed for safe chars",
			input:    "Hello. How are you? (I'm fine) [yes]",
			expected: "Hello. How are you? (I'm fine) [yes]",
		},
		{
			name:     "real example: example with exclamation",
			input:    "Hello! How are you?",
			expected: "Hello\\! How are you?",
		},
		{
			name:     "real example: part of speech with asterisk",
			input:    "interjection*",
			expected: "interjection\\*",
		},
		{
			name:     "real example: phonetic with brackets",
			input:    "[həˈləʊ]",
			expected: "[həˈləʊ]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeMarkdown(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRemoveDuplicates(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "nil slice",
			input:    nil,
			expected: []string{},
		},
		{
			name:     "single item",
			input:    []string{"hello"},
			expected: []string{"hello"},
		},
		{
			name:     "no duplicates",
			input:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "all duplicates",
			input:    []string{"a", "a", "a"},
			expected: []string{"a"},
		},
		{
			name:     "duplicates at start",
			input:    []string{"a", "a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "duplicates in middle",
			input:    []string{"a", "b", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "duplicates at end",
			input:    []string{"a", "b", "c", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "alternating duplicates",
			input:    []string{"a", "b", "a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "mixed case duplicates",
			input:    []string{"Hello", "hello", "HELLO"},
			expected: []string{"Hello", "hello", "HELLO"}, // case-sensitive
		},
		{
			name:     "empty strings",
			input:    []string{"", "a", "", "b", ""},
			expected: []string{"", "a", "b"},
		},
		{
			name:     "only empty strings",
			input:    []string{"", "", ""},
			expected: []string{""},
		},
		{
			name:     "real example: synonyms",
			input:    []string{"hi", "hey", "hello", "hi", "howdy", "hey"},
			expected: []string{"hi", "hey", "hello", "howdy"},
		},
		{
			name:     "real example: alternative translations",
			input:    []string{"привет", "здарова", "здравствуйте", "привет", "приветик"},
			expected: []string{"привет", "здарова", "здравствуйте", "приветик"},
		},
		{
			name:     "numbers as strings",
			input:    []string{"1", "2", "1", "3", "2"},
			expected: []string{"1", "2", "3"},
		},
		{
			name:     "special characters",
			input:    []string{"a!", "b@", "a!", "c#"},
			expected: []string{"a!", "b@", "c#"},
		},
		{
			name:     "whitespace strings",
			input:    []string{"a", " ", "a", "\t", " ", "b"},
			expected: []string{"a", " ", "\t", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeDuplicates(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWordS_AddWord(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx  context.Context
		word models.WordCard
	}
	tests := []struct {
		name    string
		args    args
		f       func(*mock_service.MockRepositoryI, *mock_service.MockAPII)
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				word: models.WordCard{
					UserID:      1,
					WordText:    "hello",
					Translation: "привет",
					Known:       true,
				},
			},
			f: func(mri *mock_service.MockRepositoryI, ma *mock_service.MockAPII) {
				mri.EXPECT().AddWord(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "repository error",
			args: args{
				ctx: context.Background(),
				word: models.WordCard{
					UserID:      1,
					WordText:    "hello",
					Translation: "привет",
					Known:       true,
				},
			},
			f: func(mri *mock_service.MockRepositoryI, ma *mock_service.MockAPII) {
				mri.EXPECT().AddWord(gomock.Any(), gomock.Any()).Return(errors.New("repository error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			wordService := newWordServiceMock(t, ctrl, tt.f)

			err := wordService.AddWord(tt.args.ctx, tt.args.word)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestWordS_Words(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	dateStr := now.Format("2006-01-02")

	type args struct {
		ctx     context.Context
		userID  int64
		page    int
		learned bool
	}
	tests := []struct {
		name    string
		args    args
		f       func(*mock_service.MockRepositoryI, *mock_service.MockAPII)
		want    string
		want1   bool
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx:     context.Background(),
				userID:  1,
				page:    0,
				learned: true,
			},
			f: func(mri *mock_service.MockRepositoryI, ma *mock_service.MockAPII) {
				mri.EXPECT().Words(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]models.WordCard{
					{
						UserID:      1,
						WordText:    "hello",
						Translation: "привет",
						LastSeen:    now,
						Known:       true,
					},
					{
						UserID:      1,
						WordText:    "world",
						Translation: "мир",
						LastSeen:    now,
						Known:       true,
					},
				}, 2, nil)
			},
			want: fmt.Sprintf(`📚 Страница (1/1) | Всего слов (2):

1. **hello** → *привет*
   📖 last seen: %s
2. **world** → *мир*
   📖 last seen: %s`, dateStr, dateStr),
			want1: false,
		},
		{
			name: "success: one word",
			args: args{
				ctx:     context.Background(),
				userID:  1,
				page:    0,
				learned: true,
			},
			f: func(mri *mock_service.MockRepositoryI, ma *mock_service.MockAPII) {
				mri.EXPECT().Words(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]models.WordCard{
					{
						WordText:    "cat",
						Translation: "кот",
						LastSeen:    now,
					},
				}, 1, nil)
			},
			wantErr: false,
			want: fmt.Sprintf(`📚 Страница (1/1) | Всего слов (1):

1. **cat** → *кот*
   📖 last seen: %s`, dateStr),
			want1: false,
		},
		{
			name: "success: page 1 of 2 (15 words)",
			args: args{
				ctx:     context.Background(),
				userID:  1,
				page:    1,
				learned: true,
			},
			f: func(mri *mock_service.MockRepositoryI, ma *mock_service.MockAPII) {
				mri.EXPECT().Words(gomock.Any(), int64(1), 10, true).Return([]models.WordCard{
					{WordText: "apple", Translation: "яблоко", LastSeen: now},
				}, 15, nil)
			},
			wantErr: false,
			want: fmt.Sprintf(`📚 Страница (2/2) | Всего слов (15):

11. **apple** → *яблоко*
   📖 last seen: %s`, dateStr),
			want1: false,
		},
		{
			name: "has next page",
			args: args{
				ctx:     context.Background(),
				userID:  1,
				page:    0,
				learned: true,
			},
			f: func(mri *mock_service.MockRepositoryI, ma *mock_service.MockAPII) {
				mri.EXPECT().Words(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]models.WordCard{
					{WordText: "test", Translation: "тест", LastSeen: now},
				}, 15, nil)
			},
			wantErr: false,
			want1:   true,
			want: fmt.Sprintf(`📚 Страница (1/2) | Всего слов (15):

1. **test** → *тест*
   📖 last seen: %s`, dateStr),
		},
		{
			name: "error: empty list",
			args: args{
				ctx:     context.Background(),
				userID:  1,
				page:    0,
				learned: true,
			},
			f: func(mri *mock_service.MockRepositoryI, ma *mock_service.MockAPII) {
				mri.EXPECT().Words(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]models.WordCard{}, 0, nil)
			},
			wantErr: true,
			want:    "",
			want1:   false,
		},
		{
			name: "repo error",
			args: args{
				ctx:     context.Background(),
				userID:  1,
				page:    0,
				learned: true,
			},
			f: func(mri *mock_service.MockRepositoryI, ma *mock_service.MockAPII) {
				mri.EXPECT().Words(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]models.WordCard{}, 0, errors.New("db error"))
			},
			wantErr: true,
			want:    "",
			want1:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			wordService := newWordServiceMock(t, ctrl, tt.f)

			got, got1, err := wordService.Words(tt.args.ctx, tt.args.userID, tt.args.page, tt.args.learned)
			if tt.wantErr {
				require.Error(t, err)
				assert.Empty(t, got)
				assert.False(t, got1)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want1, got1)
		})
	}
}

func TestWordS_WordStat(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx    context.Context
		userID int64
	}
	tests := []struct {
		name    string
		args    args
		f       func(*mock_service.MockRepositoryI, *mock_service.MockAPII)
		want    string
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx:    context.Background(),
				userID: 1,
			},
			f: func(mri *mock_service.MockRepositoryI, ma *mock_service.MockAPII) {
				mri.EXPECT().WordStat(gomock.Any(), gomock.Any()).Return(models.WordStats{
					TotalCount:     10,
					LearnedCount:   5,
					UnlearnedCount: 5,
				}, nil)
			},
			want: `📚 *Всего отмечено слов*: **10**

📚 *Выучено*: **5**

📚 *Предстоит запомнить*: **5**`,
			wantErr: false,
		},
		{
			name: "success: zero total",
			args: args{
				ctx:    context.Background(),
				userID: 1,
			},
			f: func(mri *mock_service.MockRepositoryI, ma *mock_service.MockAPII) {
				mri.EXPECT().WordStat(gomock.Any(), gomock.Any()).Return(models.WordStats{
					TotalCount:     0,
					LearnedCount:   0,
					UnlearnedCount: 0,
				}, nil)
			},
			want: `📚 *Всего отмечено слов*: **0**

📚 *Выучено*: **0**

📚 *Предстоит запомнить*: **0**`,
			wantErr: false,
		},
		{
			name: "success: zero learned",
			args: args{
				ctx:    context.Background(),
				userID: 1,
			},
			f: func(mri *mock_service.MockRepositoryI, ma *mock_service.MockAPII) {
				mri.EXPECT().WordStat(gomock.Any(), gomock.Any()).Return(models.WordStats{
					TotalCount:     10,
					LearnedCount:   0,
					UnlearnedCount: 10,
				}, nil)
			},
			want: `📚 *Всего отмечено слов*: **10**

📚 *Выучено*: **0**

📚 *Предстоит запомнить*: **10**`,
			wantErr: false,
		},
		{
			name: "success: all learned",
			args: args{
				ctx:    context.Background(),
				userID: 1,
			},
			f: func(mri *mock_service.MockRepositoryI, ma *mock_service.MockAPII) {
				mri.EXPECT().WordStat(gomock.Any(), gomock.Any()).Return(models.WordStats{
					TotalCount:     10,
					LearnedCount:   10,
					UnlearnedCount: 0,
				}, nil)
			},
			want: `📚 *Всего отмечено слов*: **10**

📚 *Выучено*: **10**

📚 *Предстоит запомнить*: **0**`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			wordService := newWordServiceMock(t, ctrl, tt.f)

			got, err := wordService.WordStat(tt.args.ctx, tt.args.userID)
			if tt.wantErr {
				require.Error(t, err)
				assert.Empty(t, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
