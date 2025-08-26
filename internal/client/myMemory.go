package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/DanRulev/vocabot.git/internal/models"
)

type MyMemoryAPI struct{}

func NewMyMemoryAPI() *MyMemoryAPI {
	return &MyMemoryAPI{}
}

func (m *MyMemoryAPI) TranslateEnToRu(ctx context.Context, text string) (models.MyMemoryTranslationResult, error) {
	url := fmt.Sprintf(
		"https://api.mymemory.translated.net/get?q=%s&langpair=en|ru",
		url.QueryEscape(text),
	)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return models.MyMemoryTranslationResult{}, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return models.MyMemoryTranslationResult{}, err
	}
	defer resp.Body.Close()

	var data models.MyMemoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return models.MyMemoryTranslationResult{}, err
	}

	if data.ResponseBody.ResponseStatus != 200 {
		return models.MyMemoryTranslationResult{
			Error: data.ResponseBody.ResponseDetails,
		}, nil
	}

	var alternatives []string
	for _, m := range data.Matches {
		if m.Translation != data.ResponseBody.TranslatedText {
			alternatives = append(alternatives, m.Translation)
		}
	}

	return models.MyMemoryTranslationResult{
		Text:         data.ResponseBody.TranslatedText,
		Match:        data.ResponseBody.Match,
		Source:       "en",
		Target:       "ru",
		Reliable:     data.ResponseBody.Match >= 0.8,
		Alternatives: alternatives,
	}, nil
}
