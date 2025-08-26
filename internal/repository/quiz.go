package repository

import (
	"context"

	"github.com/DanRulev/vocabot.git/internal/models"
)

type QuizR struct {
	db QueryI
}

func NewQuizRepository(db QueryI) *QuizR {
	return &QuizR{
		db: db,
	}
}

func (q *QuizR) AddQuizResult(ctx context.Context, result models.QuizCard) error {
	query := `
        INSERT INTO user_quiz_results (user_id, word, translation, type, is_correct)
        VALUES ($1, $2, $3, $4, $5)
    `

	_, err := q.db.ExecContext(ctx, query, result.UserID, result.Word, result.Translation, result.Type, result.IsCorrect)
	if err != nil {
		return err
	}

	return nil
}

func (q *QuizR) QuizStats(ctx context.Context, userID int64) (models.QuizStats, error) {
	query := `SELECT 
		COUNT(*) AS total_count,
		COALESCE(SUM(CASE WHEN is_correct THEN 1 ELSE 0 END), 0) AS right_count
	FROM user_quiz_results
	WHERE user_id = $1`

	var stats models.QuizStats
	err := q.db.GetContext(ctx, &stats, query, userID)
	if err != nil {
		return models.QuizStats{}, err
	}

	stats.WrongCount = stats.TotalCount - stats.RightCount

	return stats, nil
}
