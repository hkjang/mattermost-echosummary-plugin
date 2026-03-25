package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content json.RawMessage `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

func buildChatCompletionsURL(baseURL string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return "", errors.Wrap(err, "invalid vLLM base URL")
	}

	path := strings.TrimRight(parsed.Path, "/")
	switch {
	case strings.HasSuffix(path, "/chat/completions"):
	case strings.HasSuffix(path, "/v1"):
		parsed.Path = path + "/chat/completions"
	default:
		parsed.Path = strings.TrimRight(path, "/") + "/v1/chat/completions"
	}

	return parsed.String(), nil
}

func (p *Plugin) createChatCompletion(cfg *configuration, messages []chatMessage) (string, error) {
	endpoint, err := buildChatCompletionsURL(cfg.VLLMBaseURL)
	if err != nil {
		return "", err
	}

	payload, err := json.Marshal(chatCompletionRequest{
		Model:       cfg.VLLMModel,
		Messages:    messages,
		Temperature: 0.2,
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to encode request")
	}

	httpClient := &http.Client{Timeout: time.Duration(cfg.RequestTimeoutSeconds) * time.Second}
	request, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", errors.Wrap(err, "failed to create request")
	}
	request.Header.Set("Content-Type", "application/json")
	if cfg.VLLMAPIKey != "" {
		request.Header.Set("Authorization", "Bearer "+cfg.VLLMAPIKey)
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return "", errors.Wrap(err, "failed to call vLLM")
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed to read response")
	}

	var parsed chatCompletionResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", errors.Wrapf(err, "failed to decode vLLM response: %s", truncateForLog(string(body), 400))
	}

	if response.StatusCode >= 400 {
		if parsed.Error != nil && parsed.Error.Message != "" {
			return "", fmt.Errorf("vLLM returned %d: %s", response.StatusCode, parsed.Error.Message)
		}
		return "", fmt.Errorf("vLLM returned %d", response.StatusCode)
	}

	if len(parsed.Choices) == 0 {
		return "", errors.New("vLLM response did not include any choices")
	}

	content := extractChatContent(parsed.Choices[0].Message.Content)
	if strings.TrimSpace(content) == "" {
		return "", errors.New("vLLM returned an empty message")
	}

	return content, nil
}

func extractChatContent(raw json.RawMessage) string {
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return strings.TrimSpace(text)
	}

	var parts []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &parts); err == nil {
		lines := make([]string, 0, len(parts))
		for _, part := range parts {
			if strings.TrimSpace(part.Text) != "" {
				lines = append(lines, strings.TrimSpace(part.Text))
			}
		}
		return strings.Join(lines, "\n")
	}

	return ""
}

func truncateForLog(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit] + "..."
}
