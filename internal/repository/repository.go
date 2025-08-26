package repository

import (
	"context"
	"database/sql"
)

type QueryI interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

type Repository struct {
	*WordsR
	*QuizR
}

func NewRepository(db QueryI) Repository {
	return Repository{
		WordsR: NewWordsRepository(db),
		QuizR:  NewQuizRepository(db),
	}
}
