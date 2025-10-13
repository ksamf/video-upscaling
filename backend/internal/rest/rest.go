package rest

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

func Upscale(id uuid.UUID, baseUrl string, height int, realistic string) error {
	file := strconv.Itoa(height)
	url := fmt.Sprintf("%s/upscale/%s?file=%s&real=%s", baseUrl, id, file, realistic)
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

func CreateDubbing(id uuid.UUID, baseUrl string, to string) error {
	url := fmt.Sprintf("%s/dubbing/%s?lang=%s", baseUrl, id, to)
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
