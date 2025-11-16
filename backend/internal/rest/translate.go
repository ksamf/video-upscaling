package rest

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

func TranslateSubtitles(id uuid.UUID, baseUrl string, to string) error {
	url := fmt.Sprintf("%s/translate/%s?lang=%s", baseUrl, id, to)
	client := http.Client{}
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	return nil
}
