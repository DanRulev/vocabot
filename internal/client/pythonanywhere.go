package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/DanRulev/vocabot.git/internal/models"
)

type PythonAnyWhereAPI struct{}

func NewPythonAnyWhereAPI() *PythonAnyWhereAPI {
	return &PythonAnyWhereAPI{}
}

func (m *PythonAnyWhereAPI) DictionaryData(ctx context.Context, word string) (models.TranslationResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://ftapi.pythonanywhere.com/translate?sl=en&dl=ru&text="+word, nil)
	if err != nil {
		return models.TranslationResponse{}, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return models.TranslationResponse{}, err
	}
	defer resp.Body.Close()

	var result models.TranslationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return models.TranslationResponse{}, fmt.Errorf("failed to translate word: %v", word)
	}

	return result, nil
}
