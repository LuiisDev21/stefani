package ai

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func testConfig(url string, retries int) Config {
	return Config{
		BaseURL:     url,
		Model:       "test",
		APIKey:      "test-key",
		Temperature: 0.2,
		Timeout:     2 * time.Second,
		MaxRetries:  retries,
		JSONMode:    true,
	}
}

func newClient() *Client {
	c := New(2 * time.Second)
	// Retries must not slow the suite.
	c.sleep = func(ctx context.Context, d time.Duration) error { return nil }
	return c
}

func TestChatJSONStripsCodeFences(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("authorization = %q", got)
		}
		w.Write([]byte(`{"choices":[{"message":{"content":"` + "```json\\n{\\\"ok\\\":true}\\n```" + `"}}]}`))
	}))
	defer srv.Close()

	out, err := newClient().ChatJSON(context.Background(), testConfig(srv.URL, 0), "sys", "user")
	if err != nil {
		t.Fatalf("chat: %v", err)
	}
	if string(out) != `{"ok":true}` {
		t.Fatalf("content = %q", out)
	}
}

func TestChatJSONRetriesRateLimit(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Write([]byte(`{"choices":[{"message":{"content":"{\"ok\":true}"}}]}`))
	}))
	defer srv.Close()

	out, err := newClient().ChatJSON(context.Background(), testConfig(srv.URL, 2), "s", "u")
	if err != nil {
		t.Fatalf("expected a retry to succeed: %v", err)
	}
	if !strings.Contains(string(out), `"ok":true`) {
		t.Fatalf("content = %q", out)
	}
	if calls.Load() != 2 {
		t.Fatalf("calls = %d, want 2", calls.Load())
	}
}

func TestChatJSONStopsAfterRetries(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, err := newClient().ChatJSON(context.Background(), testConfig(srv.URL, 2), "s", "u")
	if err == nil {
		t.Fatal("expected an error after exhausting retries")
	}
	if calls.Load() != 3 {
		t.Fatalf("calls = %d, want 3 (initial + 2 retries)", calls.Load())
	}
}

func TestChatJSONSurfacesAuthErrorWithoutRetrying(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	_, err := newClient().ChatJSON(context.Background(), testConfig(srv.URL, 3), "s", "u")
	if !errors.Is(err, ErrAuth) {
		t.Fatalf("error = %v, want ErrAuth", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("calls = %d, want 1", calls.Load())
	}
}

func TestChatJSONRejectsEmptyAnswer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"choices":[]}`))
	}))
	defer srv.Close()

	_, err := newClient().ChatJSON(context.Background(), testConfig(srv.URL, 0), "s", "u")
	if !errors.Is(err, ErrEmptyAnswer) {
		t.Fatalf("error = %v, want ErrEmptyAnswer", err)
	}
}

func TestChatJSONRequiresEndpointAndKey(t *testing.T) {
	if _, err := newClient().ChatJSON(context.Background(), Config{BaseURL: "", APIKey: "k"}, "s", "u"); !errors.Is(err, ErrNoEndpoint) {
		t.Fatalf("missing url error = %v", err)
	}
	if _, err := newClient().ChatJSON(context.Background(), Config{BaseURL: "https://x", APIKey: ""}, "s", "u"); !errors.Is(err, ErrNoEndpoint) {
		t.Fatalf("missing key error = %v", err)
	}
}

func TestChatJSONRespectsContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
		io.WriteString(w, `{"choices":[{"message":{"content":"{}"}}]}`)
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_, err := newClient().ChatJSON(ctx, testConfig(srv.URL, 3), "s", "u")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error = %v, want deadline exceeded", err)
	}
}

func TestBackoffIsBounded(t *testing.T) {
	for attempt := 1; attempt < 10; attempt++ {
		d := backoff(attempt)
		if d <= 0 || d > 9*time.Second {
			t.Fatalf("backoff(%d) = %v out of range", attempt, d)
		}
	}
}
