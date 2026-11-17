// Package ai is a small, provider-agnostic client for OpenAI-compatible chat
// completion APIs. It owns transport concerns only: timeouts, retries and JSON
// extraction. Prompting and validation live in the feature packages.
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

// Errors callers must distinguish.
var (
	ErrAuth        = errors.New("the AI provider rejected the credentials")
	ErrRateLimited = errors.New("the AI provider rate limited the request")
	ErrNoEndpoint  = errors.New("no AI endpoint is configured")
	ErrEmptyAnswer = errors.New("the AI provider returned an empty answer")
)

// Config is everything needed to call a compatible endpoint.
type Config struct {
	BaseURL     string
	Model       string
	APIKey      string
	Temperature float64
	Timeout     time.Duration
	MaxRetries  int
	JSONMode    bool
}

// Client performs chat completions.
type Client struct {
	http *http.Client
	// sleep is replaceable in tests so retries do not slow the suite.
	sleep func(ctx context.Context, d time.Duration) error
}

// New builds a client with the given HTTP timeout.
func New(timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 45 * time.Second
	}
	return &Client{
		http:  &http.Client{Timeout: timeout},
		sleep: sleepCtx,
	}
}

type chatRequest struct {
	Model          string        `json:"model"`
	Messages       []chatMessage `json:"messages"`
	Temperature    float64       `json:"temperature"`
	Stream         bool          `json:"stream"`
	ResponseFormat *responseFmt  `json:"response_format,omitempty"`
}

type responseFmt struct {
	Type string `json:"type"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

// ChatJSON sends a system+user prompt and returns the JSON payload of the
// answer, with markdown fences stripped if the model wrapped it.
func (c *Client) ChatJSON(ctx context.Context, cfg Config, system, user string) ([]byte, error) {
	if strings.TrimSpace(cfg.BaseURL) == "" || strings.TrimSpace(cfg.APIKey) == "" {
		return nil, ErrNoEndpoint
	}
	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")
	endpoint := cfg.BaseURL + "/chat/completions"

	body := chatRequest{
		Model:       cfg.Model,
		Temperature: cfg.Temperature,
		Messages: []chatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
	}
	if cfg.JSONMode {
		body.ResponseFormat = &responseFmt{Type: "json_object"}
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("encode AI request: %w", err)
	}

	var lastErr error
	attempts := cfg.MaxRetries + 1
	if attempts < 1 {
		attempts = 1
	}
	for attempt := 0; attempt < attempts; attempt++ {
		if attempt > 0 {
			if err := c.sleep(ctx, backoff(attempt)); err != nil {
				return nil, err
			}
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
		if err != nil {
			return nil, fmt.Errorf("build AI request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
		req.Header.Set("Accept", "application/json")

		resp, err := c.http.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			lastErr = fmt.Errorf("AI request failed: %w", err)
			continue
		}
		raw, readErr := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
		resp.Body.Close()
		if readErr != nil {
			lastErr = fmt.Errorf("read AI response: %w", readErr)
			continue
		}

		switch resp.StatusCode {
		case http.StatusOK:
			return extractContent(raw)
		case http.StatusUnauthorized, http.StatusForbidden:
			return nil, fmt.Errorf("%w (status %d)", ErrAuth, resp.StatusCode)
		case http.StatusTooManyRequests:
			lastErr = ErrRateLimited
			continue
		}
		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("AI provider error (status %d)", resp.StatusCode)
			continue
		}
		return nil, fmt.Errorf("AI provider answered %d: %s", resp.StatusCode, snippet(raw))
	}
	if lastErr == nil {
		lastErr = ErrEmptyAnswer
	}
	return nil, lastErr
}

// Ping verifies connectivity and credentials with a tiny request.
func (c *Client) Ping(ctx context.Context, cfg Config) error {
	_, err := c.ChatJSON(ctx, cfg,
		`You are a health check. Reply only with {"ok":true}.`,
		`Reply with {"ok":true} and nothing else.`)
	return err
}

func extractContent(raw []byte) ([]byte, error) {
	var parsed chatResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("AI response is not JSON: %w", err)
	}
	if parsed.Error != nil && parsed.Error.Message != "" {
		return nil, fmt.Errorf("AI provider error: %s", parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 {
		return nil, ErrEmptyAnswer
	}
	content := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if content == "" {
		return nil, ErrEmptyAnswer
	}
	return []byte(stripFences(content)), nil
}

// stripFences removes surrounding markdown code fences some models add even in
// JSON mode.
func stripFences(s string) string {
	if !strings.HasPrefix(s, "```") {
		return s
	}
	s = strings.TrimPrefix(s, "```")
	if idx := strings.Index(s, "\n"); idx >= 0 {
		s = s[idx+1:]
	}
	s = strings.TrimSuffix(strings.TrimSpace(s), "```")
	return strings.TrimSpace(s)
}

// backoff returns the wait before attempt n: exponential with jitter, capped.
func backoff(attempt int) time.Duration {
	base := 250 * time.Millisecond * time.Duration(1<<min(attempt-1, 4))
	jitter := time.Duration(rand.Int63n(int64(120 * time.Millisecond)))
	d := base + jitter
	if d > 8*time.Second {
		d = 8 * time.Second
	}
	return d
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func snippet(raw []byte) string {
	s := strings.TrimSpace(string(raw))
	if len(s) > 200 {
		return s[:200] + "..."
	}
	return s
}
