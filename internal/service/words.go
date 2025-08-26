package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/DanRulev/vocabot.git/internal/models"
	"go.uber.org/zap"
)

type WordRI interface {
	AddWord(ctx context.Context, word models.WordCard) error
	Words(ctx context.Context, userID int64, offset int, know bool) ([]models.WordCard, int, error)
	WordStat(ctx context.Context, userID int64) (models.WordStats, error)
}

type WordS struct {
	myMemory       MyMemoryAPII
	pythonAnyWhere PythonAnyWhereAPII
	vercel         VercelAPII
	repo           WordRI
	log            *zap.Logger
}

func NewWordService(api APII, repo WordRI, log *zap.Logger) *WordS {
	return &WordS{
		myMemory:       api,
		pythonAnyWhere: api,
		vercel:         api,
		repo:           repo,
		log:            log,
	}
}

func (w *WordS) RandomWord(ctx context.Context) (string, models.WordCard, error) {
	var (
		word        string
		translate   models.MyMemoryTranslationResult
		translation string
		err         error
		maxAttempts = 5
	)

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		word, err = w.vercel.RandomWord(ctx)
		if err != nil {
			w.log.Error("failed to get random word from Vercel", zap.Int("attempt", attempt), zap.Error(err))
			if attempt == maxAttempts {
				return "", models.WordCard{}, fmt.Errorf("couldn't get the word after %d attempts: %w", maxAttempts, err)
			}
			continue
		}
		if word == "" {
			w.log.Warn("empty word received", zap.Int("attempt", attempt))
			continue
		}

		translate, err = w.myMemory.TranslateEnToRu(ctx, word)
		if err != nil {
			w.log.Error("failed to translate word", zap.String("word", word), zap.Int("attempt", attempt), zap.Error(err))
			continue
		}
		if translate.Text == "" {
			w.log.Warn("empty translate word")
			continue
		}

		translation = translate.Text

		break
	}

	dictData, err := w.pythonAnyWhere.DictionaryData(ctx, word)
	if err != nil {
		w.log.Error("failed to get dictionary data for word", zap.Error(err), zap.String("word", word))
		dictData.SourceText = word
	}

	if translation == "" {
		if dictData.DestinationText == "" {
			w.log.Error("failed to get any translation for word", zap.String("word", word))
			return "", models.WordCard{}, fmt.Errorf("failed to translate word '%s'", word)
		}
		translation = dictData.DestinationText
	}

	if dictData.DestinationText == "" {
		dictData.DestinationText = translation
	}

	formatted := formatTranslation(translate, dictData)

	wordCard := models.WordCard{
		WordText:    word,
		Translation: translation,
	}

	return formatted, wordCard, nil
}

func formatTranslation(translate models.MyMemoryTranslationResult, dictData models.TranslationResponse) string {
	var sb strings.Builder

	sourceText := dictData.SourceText
	if sourceText == "" {
		sourceText = "Unknown"
	}

	sb.WriteString("📚 *Слово*: **")
	sb.WriteString(escapeMarkdown(sourceText))
	sb.WriteString("**\n\n")

	translatedText := translate.Text
	if translatedText == "" {
		if dictData.DestinationText != "" {
			translatedText = dictData.DestinationText
		} else {
			translatedText = "неизвестно"
		}
	}

	sb.WriteString("🇷🇺 *Перевод*: ")
	sb.WriteString(escapeMarkdown(translatedText))
	sb.WriteString("\n")

	if phonetic := dictData.Pronunciation.SourceTextPhonetic; phonetic != "" {
		sb.WriteString("🔤 *Произношение*: `")
		sb.WriteString(escapeMarkdown(phonetic))
		sb.WriteString("`\n\n")
	} else {
		sb.WriteString("\n")
	}

	if len(dictData.Definitions) > 0 {
		for i, def := range dictData.Definitions {
			if i > 0 {
				sb.WriteString("\n")
			}

			sb.WriteString("🔖 *")
			sb.WriteString(escapeMarkdown(def.PartOfSpeech))
			sb.WriteString("*\n")

			sb.WriteString("📖 ")
			sb.WriteString(escapeMarkdown(def.Definition))

			if def.Example != "" {
				sb.WriteString("\n💬 _")
				sb.WriteString(escapeMarkdown(def.Example))
				sb.WriteString("_")
			}

			if len(def.OtherExamples) > 0 {
				sb.WriteString("\n📎 *Другие примеры*:\n")
				for _, ex := range def.OtherExamples {
					sb.WriteString("  • `")
					sb.WriteString(escapeMarkdown(ex))
					sb.WriteString("`\n")
				}
			}

			if len(def.Synonyms) > 0 {
				sb.WriteString("🔁 *Синонимы*:\n")
				for pos, syms := range def.Synonyms {
					sb.WriteString("  ")
					sb.WriteString(escapeMarkdown(pos))
					sb.WriteString(": ")
					sb.WriteString(strings.Join(escapeSlice(syms), ", "))
					sb.WriteString("\n")
				}
			}
			sb.WriteString("\n")
		}
	} else {
		sb.WriteString("⚠️ Нет словарных данных.\n")
	}

	if len(translate.Alternatives) > 0 {
		sb.WriteString("🔄 *Альтернативные переводы*: ")
		uniqueAlts := removeDuplicates(translate.Alternatives)
		sb.WriteString(strings.Join(escapeSlice(uniqueAlts), ", "))
		sb.WriteString("\n")
	}

	if translate.Match > 0 {
		quality := "низкое"
		if translate.Match >= 0.7 {
			quality = "высокое"
		} else if translate.Match >= 0.4 {
			quality = "среднее"
		}
		sb.WriteString(fmt.Sprintf("📊 *Качество перевода*: %.1f (%s)\n", translate.Match, quality))
	}

	return strings.TrimSpace(sb.String())
}

func escapeMarkdown(text string) string {
	for _, c := range []string{"_", "*", "#", "!"} {
		text = strings.ReplaceAll(text, c, "\\"+c)
	}
	return text
}

func escapeSlice(strs []string) []string {
	result := make([]string, len(strs))
	for i, s := range strs {
		result[i] = escapeMarkdown(s)
	}
	return result
}

func removeDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0)
	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}

func (w *WordS) AddWord(ctx context.Context, word models.WordCard) error {
	return w.repo.AddWord(ctx, word)
}

func (w *WordS) Words(ctx context.Context, userID int64, page int, learned bool) (string, bool, error) {
	words, total, err := w.repo.Words(ctx, userID, page*10, learned)
	if err != nil {
		return "", false, err
	}
	if total == 0 || len(words) == 0 {
		return "", false, fmt.Errorf("empty list")
	}

	return formatWords(words, total, page, learned), (page+1)*10 < total, nil
}

func formatWords(words []models.WordCard, total, page int, know bool) string {
	var sb strings.Builder

	totalPages := total / 10
	if total%10 != 0 {
		totalPages += 1
	}

	sb.WriteString(fmt.Sprintf("📚 Страница (%d/%d) | Всего слов (%d):\n\n", page+1, totalPages, total))

	for i, word := range words {
		num := (page * 10) + i + 1
		sb.WriteString(fmt.Sprintf("%d. **%s** → *%s*\n",
			num,
			escapeMarkdown(word.WordText),
			escapeMarkdown(word.Translation),
		))

		sb.WriteString("   📖 last seen: ")
		sb.WriteString(word.LastSeen.Format(time.DateOnly))

		if i < len(words)-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func (w *WordS) WordStat(ctx context.Context, userID int64) (string, error) {
	stats, err := w.repo.WordStat(ctx, userID)
	if err != nil {
		return "", err
	}

	return formatWordStats(stats), nil
}

func formatWordStats(stats models.WordStats) string {
	var sb strings.Builder

	sb.WriteString("📚 *Всего отмечено слов*: **")
	sb.WriteString(strconv.Itoa(stats.TotalCount))
	sb.WriteString("**\n\n")

	sb.WriteString("📚 *Выучено*: **")
	sb.WriteString(strconv.Itoa(stats.LearnedCount))
	sb.WriteString("**\n\n")

	sb.WriteString("📚 *Предстоит запомнить*: **")
	sb.WriteString(strconv.Itoa(stats.UnlearnedCount))
	sb.WriteString("**")

	return sb.String()
}
