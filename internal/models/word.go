package models

import (
	"time"
)

type WordCard struct {
	UserID      int64     `db:"user_id"`
	WordText    string    `db:"word_text"`
	Translation string    `db:"translation"`
	LastSeen    time.Time `db:"last_seen"`
	Known       bool      `db:"known"`
}

type WordStats struct {
	TotalCount     int `db:"total_count"`
	LearnedCount   int `db:"learned_count"`
	UnlearnedCount int `db:"unlearned_count"`
}
