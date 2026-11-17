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

func TestOpenCodeGoSendsSessionAndUserAgent(t *testing.T) {
	var gotSession, gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotSession = r.Header.Get("x-opencode-session")
		gotUA = r.Header.Get("User-Agent")
		w.Write([]byte(`{"choices":[{"message":{"content":"{\"ok\":true}"}}]}`))
	}))
	t.Cleanup(srv.Close)

	client := New(5 * time.Second)
	_, err := client.ChatJSON(context.Background(), Config{
		Provider:  ProviderOpenCodeGo,
		SessionID: "abc123-annotate",
		BaseURL:   srv.URL,
		Model:     "deepseek-v4.1-flash",
		APIKey:    "test-key",
	}, "system", "user")
	if err != nil {
		t.Fatalf("chat: %v", err)
	}
	if gotSession != "abc123-annotate" {
		t.Fatalf("x-opencode-session = %q, want the stable session id", gotSession)
	}
	if !strings.Contains(gotUA, "STEFANI") {
		t.Fatalf("user agent = %q, want the client to identify itself", gotUA)
	}
}

func TestOtherProvidersDoNotSendSessionHeader(t *testing.T) {
	var gotSession string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotSession = r.Header.Get("x-opencode-session")
		w.Write([]byte(`{"choices":[{"message":{"content":"{\"ok\":true}"}}]}`))
	}))
	t.Cleanup(srv.Close)

	client := New(5 * time.Second)
	_, err := client.ChatJSON(context.Background(), Config{
		Provider:  "deepseek",
		SessionID: "abc123-annotate",
		BaseURL:   srv.URL,
		Model:     "deepseek-flash",
		APIKey:    "test-key",
	}, "system", "user")
	if err != nil {
		t.Fatalf("chat: %v", err)
	}
	if gotSession != "" {
		t.Fatalf("x-opencode-session = %q, want it only for OpenCode Go", gotSession)
	}
}

func TestSessionIDIsStableAndScoped(t *testing.T) {
	a := SessionID("client", "annotate")
	if a != "client-annotate" {
		t.Fatalf("session = %q", a)
	}
	if a == SessionID("client", "ping") {
		t.Fatal("scopes must differ so routing does not mix tasks")
	}
	if SessionID("", "") != "stefani-main" {
		t.Fatalf("fallback = %q", SessionID("", ""))
	}
}

func TestModelsListsAndSortsIDs(t *testing.T) {
	var gotSession string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/models" {
			http.NotFound(w, r)
			return
		}
		gotSession = r.Header.Get("x-opencode-session")
		w.Write([]byte(`{"data":[{"id":"zeta"},{"id":"alpha"},{"id":"alpha"},{"id":""}]}`))
	}))
	t.Cleanup(srv.Close)

	client := New(5 * time.Second)
	models, err := client.Models(context.Background(), Config{
		Provider:  ProviderOpenCodeGo,
		SessionID: "abc-ping",
		BaseURL:   srv.URL,
		APIKey:    "test-key",
	})
	if err != nil {
		t.Fatalf("models: %v", err)
	}
	if len(models) != 2 || models[0] != "alpha" || models[1] != "zeta" {
		t.Fatalf("models = %v, want sorted unique ids", models)
	}
	if gotSession != "abc-ping" {
		t.Fatalf("session header = %q on /models, want it too", gotSession)
	}
}

func TestModelsReportsMissingEndpoint(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	client := New(5 * time.Second)
	_, err := client.Models(context.Background(), Config{BaseURL: srv.URL, APIKey: "k"})
	if !errors.Is(err, ErrNoModelList) {
		t.Fatalf("error = %v, want ErrNoModelList", err)
	}
}

func TestModelsRejectsBadKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	t.Cleanup(srv.Close)

	client := New(5 * time.Second)
	_, err := client.Models(context.Background(), Config{BaseURL: srv.URL, APIKey: "bad"})
	if !errors.Is(err, ErrAuth) {
		t.Fatalf("error = %v, want ErrAuth", err)
	}
}
