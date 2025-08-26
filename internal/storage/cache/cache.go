package cache

import (
	"sync"

	"github.com/DanRulev/vocabot.git/internal/models"
)

type Cache struct {
	mu    sync.Mutex
	words map[int64]models.WordCard
	quiz  map[int64]models.QuizCard
}

func NewCache() *Cache {
	return &Cache{
		words: make(map[int64]models.WordCard),
		quiz:  make(map[int64]models.QuizCard),
	}
}

func (w *Cache) SetWord(userID int64, word models.WordCard) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.words[userID] = word
}

func (w *Cache) GetWord(userID int64) (models.WordCard, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	word, exists := w.words[userID]
	return word, exists
}

func (w *Cache) DeleteWord(userID int64) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.words, userID)
}

func (w *Cache) SetQuiz(userID int64, quiz models.QuizCard) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.quiz[userID] = quiz
}

func (w *Cache) GetQuiz(userID int64) (models.QuizCard, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	quiz, exists := w.quiz[userID]
	return quiz, exists
}

func (w *Cache) DeleteQuiz(userID int64) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.quiz, userID)
}
