package models

type TranslationResponse struct {
	SourceText      string `json:"source-text"`
	DestinationText string `json:"destination-text"`
	Pronunciation   struct {
		SourceTextPhonetic   string `json:"source-text-phonetic"`
		SourceTextAudio      string `json:"source-text-audio"`
		DestinationTextAudio string `json:"destination-text-audio"`
	} `json:"pronunciation"`
	Translations struct {
		AllTranslations      [][]interface{} `json:"all-translations"`
		PossibleTranslations []string        `json:"possible-translations"`
		PossibleMistakes     interface{}     `json:"possible-mistakes"`
	} `json:"translations"`
	Definitions []struct {
		PartOfSpeech  string              `json:"part-of-speech"`
		Definition    string              `json:"definition"`
		Example       string              `json:"example"`
		OtherExamples []string            `json:"other-examples"`
		Synonyms      map[string][]string `json:"synonyms"`
	} `json:"definitions"`
	SeeAlso interface{} `json:"see-also"`
}

type MyMemoryResponse struct {
	ResponseBody struct {
		TranslatedText   string  `json:"translatedText"`
		Match            float64 `json:"match"`            // 0.0 - 1.0: степень совпадения с базой
		ResponseStatus   int     `json:"responseStatus"`   // HTTP статус (200, 404 и т.д.)
		ResponseDetails  string  `json:"responseDetails"`  // "OK", "Daily request limit reached"
		ExceptionMessage string  `json:"exceptionMessage"` // если ошибка
		MTLangSupported  bool    `json:"mtLangSupported"`  // поддерживается ли машинный перевод
		TranslatedTTL    int     `json:"translatedTTL"`    // время жизни перевода (в секундах)
	} `json:"responseData"`

	// Альтернативные варианты перевода (если есть)
	Matches []struct {
		Translation string `json:"translation"` // перевод
	} `json:"matches"`
}

type MyMemoryTranslationResult struct {
	Text         string
	Match        float64
	Source       string
	Target       string
	Reliable     bool
	Alternatives []string
	Error        string
}
