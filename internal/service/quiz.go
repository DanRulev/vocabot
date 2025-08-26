package service

import (
	"context"
	crypto "crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"strconv"
	"strings"
	"sync"

	"github.com/DanRulev/vocabot.git/internal/models"
	"go.uber.org/zap"
)

type QuizRI interface {
	AddQuizResult(ctx context.Context, result models.QuizCard) error
	QuizStats(ctx context.Context, userID int64) (models.QuizStats, error)
}

type AuxiliaryWord interface {
	AddWord(ctx context.Context, word models.WordCard) error
	RandomUnknownWord(ctx context.Context, userID int64) (models.WordCard, error)
}

type QuizS struct {
	myMemory       MyMemoryAPII
	pythonAnyWhere PythonAnyWhereAPII
	vercel         VercelAPII
	repo           QuizRI
	aux            AuxiliaryWord
	log            *zap.Logger
}

func NewQuizService(api APII, repo QuizRI, aux AuxiliaryWord, log *zap.Logger) *QuizS {
	return &QuizS{
		myMemory:       api,
		pythonAnyWhere: api,
		vercel:         api,
		repo:           repo,
		aux:            aux,
		log:            log,
	}
}

func (q *QuizS) NewQuiz(ctx context.Context, userID int64) (string, map[string]bool, error) {
	var target string

	quiz := make(map[string]bool)
	used := make(map[string]bool)

	var (
		mu          sync.Mutex
		wg          sync.WaitGroup
		errs        []error
		maxAttempts = 5
	)

	truePosition, err := randomPosition(4)
	if err != nil {
		q.log.Warn("crypto/rand failed, using math/rand fallback", zap.Error(err))
		truePosition = rand.Intn(4)
	}

	for i := 0; i < 4; i++ {
		correctness := (truePosition == i)
		wg.Add(1)
		go func(correctness bool) {
			defer wg.Done()

			var translation string

			for attempts := 0; attempts < maxAttempts; attempts++ {
				word, err := q.vercel.RandomWord(ctx)
				if err != nil {
					mu.Lock()
					errs = append(errs, fmt.Errorf("RandomWord failed: %w", err))
					mu.Unlock()
					continue
				}
				if correctness {
					mu.Lock()
					target = word
					mu.Unlock()
				}

				trans, err := q.myMemory.TranslateEnToRu(ctx, word)
				if err != nil {
					mu.Lock()
					errs = append(errs, fmt.Errorf("TranslateWord failed: %w", err))
					mu.Unlock()
					continue
				}

				if trans.Text == "" {
					dictData, err := q.pythonAnyWhere.DictionaryData(ctx, word)
					if err != nil {
						mu.Lock()
						errs = append(errs, fmt.Errorf("DictionaryData failed: %w", err))
						mu.Unlock()
						continue
					}
					if dictData.DestinationText == "" {
						mu.Lock()
						errs = append(errs, fmt.Errorf("translation empty: %v", word))
						mu.Unlock()
						continue
					}
					translation = dictData.DestinationText
				} else {
					translation = trans.Text
				}

				mu.Lock()
				if !used[translation] {
					used[translation] = true
					mu.Unlock()
					break
				}
				mu.Unlock()
			}

			mu.Lock()
			quiz[translation] = correctness
			mu.Unlock()
		}(correctness)
	}

	wg.Wait()

	if len(errs) > 0 {
		q.log.Warn("errors during NewQuiz", zap.Int("error_count", len(errs)), zap.Errors("errors", errs))
	}

	if len(quiz) < 4 {
		q.log.Warn("not enough unique translations", zap.Int("got", len(quiz)), zap.Int("required", 4))
		return "", nil, errors.New("not enough unique translations")
	}

	return target, quiz, nil
}

func (q *QuizS) AddQuizResult(ctx context.Context, result models.QuizCard) error {
	err := q.aux.AddWord(ctx, models.WordCard{
		UserID:      result.UserID,
		WordText:    result.Word,
		Translation: result.Translation,
		Known:       result.IsCorrect,
	})
	if err != nil {
		q.log.Warn("failed to add word to user's dictionary", zap.Int64("user_id", result.UserID), zap.String("word", result.Word), zap.Error(err))
	}
	return q.repo.AddQuizResult(ctx, result)
}

func (q *QuizS) QuizStats(ctx context.Context, userID int64) (string, error) {
	stats, err := q.repo.QuizStats(ctx, userID)
	if err != nil {
		q.log.Warn("failed to get quiz stats", zap.Int64("user_id", userID), zap.Error(err))
		return "", err
	}

	return quizStatsFormat(stats), nil
}

func quizStatsFormat(stats models.QuizStats) string {
	var sb strings.Builder

	sb.WriteString("📚 *Всего попыток*: **")
	sb.WriteString(strconv.Itoa(stats.TotalCount))
	sb.WriteString("**\n\n")

	sb.WriteString("📚 *Удачных*: **")
	sb.WriteString(strconv.Itoa(stats.RightCount))
	sb.WriteString("**\n\n")

	sb.WriteString("📚 *Не удачных*: **")
	sb.WriteString(strconv.Itoa(stats.WrongCount))
	sb.WriteString("**")

	return sb.String()
}

func randomPosition(max int64) (int, error) {
	if max <= 0 {
		return 0, errors.New("max must be greater than 0")
	}

	n, err := crypto.Int(crypto.Reader, big.NewInt(max))
	if err != nil {
		return 0, err
	}

	return int(n.Int64()), nil
}
