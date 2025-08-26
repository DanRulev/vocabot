package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DanRulev/vocabot.git/internal/models"
	mock_repository "github.com/DanRulev/vocabot.git/internal/repository/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newWordsMock(t *testing.T, ctrl *gomock.Controller, setupMock func(*mock_repository.MockQueryI)) *WordsR {
	db := mock_repository.NewMockQueryI(ctrl)
	if setupMock != nil {
		setupMock(db)
	}

	return &WordsR{db: db}
}

func TestWordsR_AddWord(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx  context.Context
		word models.WordCard
	}
	tests := []struct {
		name    string
		args    args
		f       func(*mock_repository.MockQueryI)
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx:  context.Background(),
				word: models.WordCard{},
			},
			f: func(mqi *mock_repository.MockQueryI) {
				mqi.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)
			},
			wantErr: false,
		},
		{
			name: "error exec",
			args: args{
				ctx:  context.Background(),
				word: models.WordCard{},
			},
			f: func(mqi *mock_repository.MockQueryI) {
				mqi.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("error exec"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := newWordsMock(t, ctrl, tt.f)

			err := repo.AddWord(tt.args.ctx, tt.args.word)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestWordsR_RandomUnknownWord(t *testing.T) {
	t.Parallel()

	randomWord := models.WordCard{
		UserID:      1,
		WordText:    "example",
		Translation: "пример",
		LastSeen:    time.Now(),
		Known:       false,
	}

	type args struct {
		ctx    context.Context
		userID int64
	}
	tests := []struct {
		name    string
		args    args
		f       func(*mock_repository.MockQueryI)
		want    models.WordCard
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx:    context.Background(),
				userID: 1,
			},
			f: func(mqi *mock_repository.MockQueryI) {
				mqi.EXPECT().GetContext(gomock.Any(), gomock.AssignableToTypeOf(&randomWord), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
						*dest.(*models.WordCard) = randomWord
						return nil
					})
			},
			want:    randomWord,
			wantErr: false,
		},
		{
			name: "failed no rows",
			args: args{
				ctx:    context.Background(),
				userID: 1,
			},
			f: func(mqi *mock_repository.MockQueryI) {
				mqi.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(sql.ErrNoRows)
			},
			want:    models.WordCard{},
			wantErr: true,
		},
		{
			name: "db error",
			args: args{
				ctx:    context.Background(),
				userID: 1,
			},
			f: func(mqi *mock_repository.MockQueryI) {
				mqi.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("db error"))
			},
			want:    models.WordCard{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := newWordsMock(t, ctrl, tt.f)

			got, err := repo.RandomUnknownWord(tt.args.ctx, tt.args.userID)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, got.UserID, tt.want.UserID)
			assert.Equal(t, got.WordText, tt.want.WordText)
			assert.Equal(t, got.Translation, tt.want.Translation)
			assert.Equal(t, got.Known, tt.want.Known)
			assert.WithinDuration(t, got.LastSeen, tt.want.LastSeen, time.Second)
		})
	}
}

func TestWordsR_Words(t *testing.T) {
	t.Parallel()

	expectedWords := []models.WordCard{{UserID: 1, WordText: "hello", Translation: "привет", Known: true}}

	type args struct {
		ctx    context.Context
		userID int64
		offset int
		known  bool
	}
	tests := []struct {
		name    string
		args    args
		f       func(*mock_repository.MockQueryI)
		want1   []models.WordCard
		want2   int
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx:    context.Background(),
				userID: 1,
				offset: 0,
				known:  true,
			},
			f: func(mqi *mock_repository.MockQueryI) {
				var total int
				mqi.EXPECT().GetContext(gomock.Any(), gomock.AssignableToTypeOf(&total), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
						*dest.(*int) = 1
						return nil
					})

				mqi.EXPECT().SelectContext(gomock.Any(), gomock.AssignableToTypeOf(&expectedWords), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
						slice := dest.(*[]models.WordCard)
						*slice = append(*slice, expectedWords...)
						return nil
					})
			},
			want1:   expectedWords,
			want2:   1,
			wantErr: false,
		},
		{
			name: "failed get total",
			args: args{
				ctx:    context.Background(),
				userID: 1,
				offset: 0,
				known:  true,
			},
			f: func(mqi *mock_repository.MockQueryI) {
				var total int
				mqi.EXPECT().GetContext(gomock.Any(), gomock.AssignableToTypeOf(&total), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
						*dest.(*int) = 0
						return errors.New("failed get total")
					})
			},
			want1:   []models.WordCard{},
			want2:   0,
			wantErr: true,
		},
		{
			name: "failed 0 total count",
			args: args{
				ctx:    context.Background(),
				userID: 1,
				offset: 0,
				known:  true,
			},
			f: func(mqi *mock_repository.MockQueryI) {
				var total int
				mqi.EXPECT().GetContext(gomock.Any(), gomock.AssignableToTypeOf(&total), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
						*dest.(*int) = 0
						return nil
					})
			},
			want1:   []models.WordCard{},
			want2:   0,
			wantErr: false,
		},
		{
			name: "db error",
			args: args{
				ctx:    context.Background(),
				userID: 1,
				offset: 0,
				known:  true,
			},
			f: func(mqi *mock_repository.MockQueryI) {
				var total int
				mqi.EXPECT().GetContext(gomock.Any(), gomock.AssignableToTypeOf(&total), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
						*dest.(*int) = 1
						return nil
					})

				mqi.EXPECT().SelectContext(gomock.Any(), gomock.AssignableToTypeOf(&expectedWords), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
						return errors.New("db error")
					})
			},
			want1:   nil,
			want2:   0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := newWordsMock(t, ctrl, tt.f)

			got1, got2, err := repo.Words(tt.args.ctx, tt.args.userID, tt.args.offset, tt.args.known)
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, got2, tt.want2)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want1, got1)
			assert.Equal(t, tt.want2, got2)
		})
	}
}

func TestWordsR_WordStat(t *testing.T) {
	t.Parallel()

	expectedStats := models.WordStats{
		TotalCount:     10,
		LearnedCount:   5,
		UnlearnedCount: 5,
	}

	type args struct {
		ctx    context.Context
		userID int64
	}
	tests := []struct {
		name    string
		args    args
		f       func(*mock_repository.MockQueryI)
		want    models.WordStats
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx:    context.Background(),
				userID: 1,
			},
			f: func(mqi *mock_repository.MockQueryI) {
				mqi.EXPECT().GetContext(gomock.Any(), gomock.AssignableToTypeOf(&expectedStats), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
						*dest.(*models.WordStats) = expectedStats
						return nil
					})
			},
			want:    expectedStats,
			wantErr: false,
		},
		{
			name: "db error",
			args: args{
				ctx:    context.Background(),
				userID: 1,
			},
			f: func(mqi *mock_repository.MockQueryI) {
				mqi.EXPECT().GetContext(gomock.Any(), gomock.AssignableToTypeOf(&expectedStats), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
						return errors.New("db error")
					})
			},
			want:    models.WordStats{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := newWordsMock(t, ctrl, tt.f)

			got, err := repo.WordStat(tt.args.ctx, tt.args.userID)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, got.TotalCount, tt.want.TotalCount)
			assert.Equal(t, got.LearnedCount, tt.want.LearnedCount)
			assert.Equal(t, got.UnlearnedCount, tt.want.UnlearnedCount)
		})
	}
}
