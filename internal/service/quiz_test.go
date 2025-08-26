package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/DanRulev/vocabot.git/internal/models"
	mock_service "github.com/DanRulev/vocabot.git/internal/service/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newQuizServiceMock(t *testing.T, ctrl *gomock.Controller, setupMock func(*mock_service.MockRepositoryI, *mock_service.MockAPII)) *QuizS {
	api := mock_service.NewMockAPII(ctrl)
	repo := mock_service.NewMockRepositoryI(ctrl)
	if setupMock != nil {
		setupMock(repo, api)
	}

	log := zap.NewNop()

	return &QuizS{
		pythonAnyWhere: api,
		myMemory:       api,
		vercel:         api,
		repo:           repo,
		aux:            repo,
		log:            log,
	}
}

func TestQuizS_NewQuiz(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx    context.Context
		userID int64
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
				ctx:    context.Background(),
				userID: 1,
			},
			f: func(mri *mock_service.MockRepositoryI, ma *mock_service.MockAPII) {
				ma.EXPECT().RandomWord(gomock.Any()).Return("hello", nil)
				ma.EXPECT().RandomWord(gomock.Any()).Return("home", nil)
				ma.EXPECT().RandomWord(gomock.Any()).Return("sun", nil)
				ma.EXPECT().RandomWord(gomock.Any()).Return("night", nil)

				ma.EXPECT().TranslateEnToRu(gomock.Any(), "hello").Return(models.MyMemoryTranslationResult{
					Text: "привет",
				}, nil)
				ma.EXPECT().TranslateEnToRu(gomock.Any(), "home").Return(models.MyMemoryTranslationResult{
					Text: "дом",
				}, nil)
				ma.EXPECT().TranslateEnToRu(gomock.Any(), "sun").Return(models.MyMemoryTranslationResult{
					Text: "солнце",
				}, nil)
				ma.EXPECT().TranslateEnToRu(gomock.Any(), "night").Return(models.MyMemoryTranslationResult{
					Text: "ночь",
				}, nil)

			},
			wantErr: false,
		},
		{
			name: "success: vercel.RandomWord returns error",
			args: args{
				ctx:    context.Background(),
				userID: 1,
			},
			f: func(mri *mock_service.MockRepositoryI, ma *mock_service.MockAPII) {
				ma.EXPECT().RandomWord(gomock.Any()).Return("", errors.New("service unavailable"))
				ma.EXPECT().RandomWord(gomock.Any()).Return("hello", nil)
				ma.EXPECT().RandomWord(gomock.Any()).Return("home", nil)
				ma.EXPECT().RandomWord(gomock.Any()).Return("sun", nil)
				ma.EXPECT().RandomWord(gomock.Any()).Return("night", nil)

				ma.EXPECT().TranslateEnToRu(gomock.Any(), "hello").Return(models.MyMemoryTranslationResult{
					Text: "привет",
				}, nil)
				ma.EXPECT().TranslateEnToRu(gomock.Any(), "home").Return(models.MyMemoryTranslationResult{
					Text: "дом",
				}, nil)
				ma.EXPECT().TranslateEnToRu(gomock.Any(), "sun").Return(models.MyMemoryTranslationResult{
					Text: "солнце",
				}, nil)
				ma.EXPECT().TranslateEnToRu(gomock.Any(), "night").Return(models.MyMemoryTranslationResult{
					Text: "ночь",
				}, nil)
			},
			wantErr: false,
		},
		{
			name: "success: pythonAnyWhere.TranslateEnToRu returns error",
			args: args{
				ctx:    context.Background(),
				userID: 1,
			},
			f: func(mri *mock_service.MockRepositoryI, ma *mock_service.MockAPII) {
				ma.EXPECT().RandomWord(gomock.Any()).Return("hello", nil)
				ma.EXPECT().RandomWord(gomock.Any()).Return("home", nil)
				ma.EXPECT().RandomWord(gomock.Any()).Return("sun", nil)
				ma.EXPECT().RandomWord(gomock.Any()).Return("night", nil)
				ma.EXPECT().RandomWord(gomock.Any()).Return("night", nil)

				ma.EXPECT().TranslateEnToRu(gomock.Any(), "hello").Return(models.MyMemoryTranslationResult{
					Text: "привет",
				}, nil)
				ma.EXPECT().TranslateEnToRu(gomock.Any(), "home").Return(models.MyMemoryTranslationResult{
					Text: "дом",
				}, nil)
				ma.EXPECT().TranslateEnToRu(gomock.Any(), "sun").Return(models.MyMemoryTranslationResult{
					Text: "солнце",
				}, nil)
				ma.EXPECT().TranslateEnToRu(gomock.Any(), "night").Return(models.MyMemoryTranslationResult{
					Text: "",
				}, errors.New("service unavailable"))
				ma.EXPECT().TranslateEnToRu(gomock.Any(), "night").Return(models.MyMemoryTranslationResult{
					Text: "ночь",
				}, nil)
			},
			wantErr: false,
		},
		{
			name: "success: skip empty translations and retry",
			args: args{
				ctx:    context.Background(),
				userID: 1,
			},
			f: func(mri *mock_service.MockRepositoryI, ma *mock_service.MockAPII) {
				ma.EXPECT().RandomWord(gomock.Any()).Return("hello", nil)
				ma.EXPECT().RandomWord(gomock.Any()).Return("home", nil)
				ma.EXPECT().RandomWord(gomock.Any()).Return("sun", nil)
				ma.EXPECT().RandomWord(gomock.Any()).Return("night", nil)
				ma.EXPECT().RandomWord(gomock.Any()).Return("night", nil)

				ma.EXPECT().TranslateEnToRu(gomock.Any(), "hello").Return(models.MyMemoryTranslationResult{
					Text: "привет",
				}, nil)
				ma.EXPECT().TranslateEnToRu(gomock.Any(), "home").Return(models.MyMemoryTranslationResult{
					Text: "дом",
				}, nil)
				ma.EXPECT().TranslateEnToRu(gomock.Any(), "sun").Return(models.MyMemoryTranslationResult{
					Text: "солнце",
				}, nil)
				ma.EXPECT().TranslateEnToRu(gomock.Any(), "night").Return(models.MyMemoryTranslationResult{}, nil)
				ma.EXPECT().DictionaryData(gomock.Any(), gomock.Any()).Return(models.TranslationResponse{}, nil)
				ma.EXPECT().TranslateEnToRu(gomock.Any(), "night").Return(models.MyMemoryTranslationResult{
					Text: "ночь",
				}, nil)

			},
			wantErr: false,
		},
		{
			name: "success: avoid duplicate translations",
			args: args{
				ctx:    context.Background(),
				userID: 1,
			},
			f: func(mri *mock_service.MockRepositoryI, ma *mock_service.MockAPII) {
				ma.EXPECT().RandomWord(gomock.Any()).Return("hello", nil)
				ma.EXPECT().RandomWord(gomock.Any()).Return("hello", nil)
				ma.EXPECT().RandomWord(gomock.Any()).Return("home", nil)
				ma.EXPECT().RandomWord(gomock.Any()).Return("sun", nil)
				ma.EXPECT().RandomWord(gomock.Any()).Return("night", nil)

				ma.EXPECT().TranslateEnToRu(gomock.Any(), "hello").Return(models.MyMemoryTranslationResult{
					Text: "привет",
				}, nil)
				ma.EXPECT().TranslateEnToRu(gomock.Any(), "hello").Return(models.MyMemoryTranslationResult{
					Text: "привет",
				}, nil)
				ma.EXPECT().TranslateEnToRu(gomock.Any(), "home").Return(models.MyMemoryTranslationResult{
					Text: "дом",
				}, nil)
				ma.EXPECT().TranslateEnToRu(gomock.Any(), "sun").Return(models.MyMemoryTranslationResult{
					Text: "солнце",
				}, nil)
				ma.EXPECT().TranslateEnToRu(gomock.Any(), "night").Return(models.MyMemoryTranslationResult{
					Text: "ночь",
				}, nil)
			},
			wantErr: false,
		},
		{
			name: "error: RandomWord fails",
			args: args{
				ctx:    context.Background(),
				userID: 1,
			},
			f: func(mri *mock_service.MockRepositoryI, ma *mock_service.MockAPII) {
				ma.EXPECT().RandomWord(gomock.Any()).Return("", errors.New("service down")).Times(20)
			},
			wantErr: true,
		},
		{
			name: "error: RandomWord fails",
			args: args{
				ctx:    context.Background(),
				userID: 1,
			},
			f: func(mri *mock_service.MockRepositoryI, ma *mock_service.MockAPII) {
				ma.EXPECT().RandomWord(gomock.Any()).Return("", errors.New("service down")).Times(20)
			},
			wantErr: true,
		},
		{
			name: "error: TranslateEnToRu fails",
			args: args{
				ctx:    context.Background(),
				userID: 1,
			},
			f: func(mri *mock_service.MockRepositoryI, ma *mock_service.MockAPII) {
				ma.EXPECT().RandomWord(gomock.Any()).Return("hello", nil).Times(20)
				ma.EXPECT().TranslateEnToRu(gomock.Any(), "hello").Return(models.MyMemoryTranslationResult{}, errors.New("translation failed")).Times(20)
			},
			wantErr: true,
		},
		{
			name: "error: empty translates3",
			args: args{
				ctx:    context.Background(),
				userID: 1,
			},
			f: func(mri *mock_service.MockRepositoryI, ma *mock_service.MockAPII) {
				ma.EXPECT().RandomWord(gomock.Any()).Return("hello", nil).Times(20)
				ma.EXPECT().TranslateEnToRu(gomock.Any(), gomock.Any()).Return(models.MyMemoryTranslationResult{}, nil).Times(20)
				ma.EXPECT().DictionaryData(gomock.Any(), gomock.Any()).Return(models.TranslationResponse{}, nil).Times(20)
			},
			wantErr: true,
		},
		{
			name: "error: not enough unique translations",
			args: args{
				ctx:    context.Background(),
				userID: 1,
			},
			f: func(mri *mock_service.MockRepositoryI, ma *mock_service.MockAPII) {
				for i := range 16 {
					ma.EXPECT().RandomWord(gomock.Any()).Return(fmt.Sprintf("hello%d", i), nil)

				}
				ma.EXPECT().TranslateEnToRu(gomock.Any(), gomock.Any()).Return(models.MyMemoryTranslationResult{
					Text: "привет",
				}, nil).Times(16)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			quizService := newQuizServiceMock(t, ctrl, tt.f)

			target, quiz, err := quizService.NewQuiz(tt.args.ctx, tt.args.userID)

			if tt.wantErr {
				require.Error(t, err)
				assert.Empty(t, target)
				assert.Nil(t, quiz)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, target)
			assert.Len(t, quiz, 4)

			var trueCnt, falseCnt int
			for _, v := range quiz {
				if v {
					trueCnt++
				} else {
					falseCnt++
				}
			}

			assert.Equal(t, trueCnt, 1)
			assert.Equal(t, falseCnt, 3)
		})
	}
}

func TestQuizS_AddQuizResult(t *testing.T) {
	t.Parallel()

	result := models.QuizCard{
		UserID:      1,
		Word:        "hello",
		Translation: "привет",
		Type:        "quiz",
		IsCorrect:   true,
	}

	type args struct {
		ctx    context.Context
		result models.QuizCard
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
				ctx:    context.Background(),
				result: result,
			},
			f: func(mri *mock_service.MockRepositoryI, ma *mock_service.MockAPII) {
				mri.EXPECT().AddWord(gomock.Any(), gomock.Any()).Return(nil)
				mri.EXPECT().AddQuizResult(gomock.Any(), result).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "success: AddWord fails, but AddQuizResult succeeds",
			args: args{
				ctx:    context.Background(),
				result: result,
			},
			f: func(mri *mock_service.MockRepositoryI, ma *mock_service.MockAPII) {
				mri.EXPECT().AddWord(gomock.Any(), gomock.Any()).Return(errors.New("db error"))
				mri.EXPECT().AddQuizResult(gomock.Any(), result).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error: AddQuizResult fails",
			args: args{
				ctx:    context.Background(),
				result: result,
			},
			f: func(mri *mock_service.MockRepositoryI, ma *mock_service.MockAPII) {
				mri.EXPECT().AddWord(gomock.Any(), gomock.Any()).Return(nil)
				mri.EXPECT().AddQuizResult(gomock.Any(), result).Return(errors.New("failed to save quiz result"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			quizService := newQuizServiceMock(t, ctrl, tt.f)

			err := quizService.AddQuizResult(tt.args.ctx, tt.args.result)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestQuizS_QuizStats(t *testing.T) {
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
				mri.EXPECT().QuizStats(gomock.Any(), gomock.Any()).Return(models.QuizStats{
					TotalCount: 10,
					RightCount: 7,
					WrongCount: 3,
				}, nil)
			},
			want: `📚 *Всего попыток*: **10**

📚 *Удачных*: **7**

📚 *Не удачных*: **3**`,
			wantErr: false,
		},
		{
			name: "error: repo failure",
			args: args{
				ctx:    context.Background(),
				userID: 1,
			},
			f: func(mri *mock_service.MockRepositoryI, ma *mock_service.MockAPII) {
				mri.EXPECT().QuizStats(gomock.Any(), gomock.Any()).Return(models.QuizStats{}, errors.New("database unreachable"))
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			quizService := newQuizServiceMock(t, ctrl, tt.f)

			got, err := quizService.QuizStats(tt.args.ctx, tt.args.userID)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_randomPosition(t *testing.T) {
	t.Parallel()

	type args struct {
		max int64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "success",
			args:    args{max: 10},
			wantErr: false,
		},
		{
			name:    "success: max=1",
			args:    args{max: 1},
			wantErr: false,
		},
		{
			name:    "failed",
			args:    args{max: 0},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			n, err := randomPosition(tt.args.max)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.GreaterOrEqual(t, n, 0)
			assert.Less(t, n, int(tt.args.max))
		})
	}
}
