package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/DanRulev/vocabot.git/internal/models"
	mock_repository "github.com/DanRulev/vocabot.git/internal/repository/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newQuizMock(t *testing.T, ctrl *gomock.Controller, setupMock func(*mock_repository.MockQueryI)) *QuizR {
	db := mock_repository.NewMockQueryI(ctrl)
	if setupMock != nil {
		setupMock(db)
	}

	return &QuizR{db: db}
}

func TestQuizR_AddQuizResult(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx    context.Context
		result models.QuizCard
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
				ctx:    context.Background(),
				result: models.QuizCard{},
			},
			f: func(mqi *mock_repository.MockQueryI) {
				mqi.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)
			},
			wantErr: false,
		},
		{
			name: "failed exec",
			args: args{
				ctx:    context.Background(),
				result: models.QuizCard{},
			},
			f: func(mqi *mock_repository.MockQueryI) {
				mqi.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("exec error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			quizR := newQuizMock(t, ctrl, tt.f)

			err := quizR.AddQuizResult(tt.args.ctx, tt.args.result)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestQuizR_QuizStats(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx    context.Context
		userID int64
	}
	tests := []struct {
		name    string
		args    args
		f       func(*mock_repository.MockQueryI)
		want    models.QuizStats
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx:    context.Background(),
				userID: 1,
			},
			f: func(mqi *mock_repository.MockQueryI) {
				mqi.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			want: models.QuizStats{
				TotalCount: 0,
				RightCount: 0,
				WrongCount: 0,
			},
			wantErr: false,
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
			want:    models.QuizStats{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			quizR := newQuizMock(t, ctrl, tt.f)

			got, err := quizR.QuizStats(tt.args.ctx, tt.args.userID)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, got.TotalCount, tt.want.TotalCount)
			assert.Equal(t, got.RightCount, tt.want.RightCount)
			assert.Equal(t, got.WrongCount, tt.want.WrongCount)
		})
	}
}
