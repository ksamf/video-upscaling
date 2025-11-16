package rest

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

func Upscale(id uuid.UUID, baseUrl string, height int, realistic bool) error {
	file := strconv.Itoa(height)
	url := fmt.Sprintf("%s/upscale/%s?file=%s&real=%t", baseUrl, id, file, realistic)
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
