package rest

import (
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
)

func CreateSubtitles(id uuid.UUID, baseUrl string) (string, error) {
	url := fmt.Sprintf("%s/subtitles/%s", baseUrl, id)
	client := http.Client{}
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read body:%w", err)
	}
	lang := string(body)
	defer resp.Body.Close()
	return lang, nil
}
