package service

import (
	"context"

	"github.com/DanRulev/vocabot.git/internal/models"
	"go.uber.org/zap"
)

type MyMemoryAPII interface {
	TranslateEnToRu(ctx context.Context, text string) (models.MyMemoryTranslationResult, error)
}

type PythonAnyWhereAPII interface {
	DictionaryData(ctx context.Context, word string) (models.TranslationResponse, error)
}

type VercelAPII interface {
	RandomWord(ctx context.Context) (string, error)
}

type APII interface {
	MyMemoryAPII
	PythonAnyWhereAPII
	VercelAPII
}

type RepositoryI interface {
	AuxiliaryWord
	QuizRI
	WordRI
}

type Service struct {
	*WordS
	*QuizS
}

func InitServices(api APII, repo RepositoryI, log *zap.Logger) *Service {
	return &Service{
		WordS: NewWordService(api, repo, log),
		QuizS: NewQuizService(api, repo, repo, log),
	}
}
