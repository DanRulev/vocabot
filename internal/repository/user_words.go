package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DanRulev/vocabot.git/internal/models"
)

type WordsR struct {
	db QueryI
}

func NewWordsRepository(db QueryI) *WordsR {
	return &WordsR{db: db}
}

func (w *WordsR) AddWord(ctx context.Context, word models.WordCard) error {
	query := `INSERT INTO user_words (user_id, word_text, translation, known, last_seen)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (user_id, word_text)
		DO UPDATE SET
			known = CASE 
				WHEN EXCLUDED.known THEN true  
				ELSE user_words.known            
			END,
			last_seen = NOW()
		`
	_, err := w.db.ExecContext(ctx, query, word.UserID, word.WordText, word.Translation, word.Known)
	if err != nil {
		return err
	}

	return nil
}

func (w *WordsR) RandomUnknownWord(ctx context.Context, userID int64) (models.WordCard, error) {
	query := `
	SELECT word_text, translation
		FROM user_words
		WHERE user_id = $1 AND known = false
		ORDER BY RANDOM()
		LIMIT 1;
	`

	var word models.WordCard
	err := w.db.GetContext(ctx, &word, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.WordCard{}, fmt.Errorf("no unknown words found for user %d", userID)
		}
		return models.WordCard{}, fmt.Errorf("database error: %w", err)
	}
	return word, nil
}

func (w *WordsR) Words(ctx context.Context, userID int64, offset int, known bool) ([]models.WordCard, int, error) {
	var total int
	countQuery := `SELECT COUNT(*) FROM user_words WHERE user_id = $1 AND known = $2`
	err := w.db.GetContext(ctx, &total, countQuery, userID, known)
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []models.WordCard{}, 0, nil
	}

	query := `
		SELECT user_id, word_text, translation, last_seen, known
		FROM user_words
		WHERE user_id = $1 AND known = $2
		ORDER BY last_seen DESC
		LIMIT 10 OFFSET $3
	`
	words := make([]models.WordCard, 0, 10)
	err = w.db.SelectContext(ctx, &words, query, userID, known, offset)
	if err != nil {
		return nil, 0, err
	}

	return words, total, nil
}

func (w *WordsR) WordStat(ctx context.Context, userID int64) (models.WordStats, error) {
	query := `
		SELECT
			COUNT(*) AS total_count,
			COALESCE(SUM(CASE WHEN known THEN 1 ELSE 0 END), 0) AS learned_count
		FROM user_words
		WHERE user_id = $1
	`

	var stats models.WordStats
	err := w.db.GetContext(ctx, &stats, query, userID)
	if err != nil {
		return models.WordStats{}, fmt.Errorf("failed to get word stats for user %d: %w", userID, err)
	}

	stats.UnlearnedCount = stats.TotalCount - stats.LearnedCount

	return stats, nil
}
