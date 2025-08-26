package models

import "time"

type QuizCard struct {
	UserID      int64     `db:"user_id"`
	Word        string    `db:"word"`
	Translation string    `db:"translation"`
	Type        string    `db:"type"`
	IsCorrect   bool      `db:"is_correct"`
	LastSeen    time.Time `db:"last_seen"`
}

type QuizStats struct {
	TotalCount int `db:"total_count"`
	RightCount int `db:"right_count"`
	WrongCount int `db:"wrong_count"`
}
