package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

type VercelAPI struct{}

func NewVercelAPI() *VercelAPI {
	return &VercelAPI{}
}

func (v *VercelAPI) RandomWord(ctx context.Context) (string, error) {
	resp, err := http.Get("https://random-word-api.vercel.app/api?words=1")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var word []string
	if err = json.NewDecoder(resp.Body).Decode(&word); err != nil {
		return "", errors.New("failed to get random word")
	}

	return word[0], nil
}
