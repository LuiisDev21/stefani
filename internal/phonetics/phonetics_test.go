package phonetics

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/LuiisDev21/stefani/internal/ai"
	"github.com/LuiisDev21/stefani/internal/config"
	"github.com/LuiisDev21/stefani/internal/domain"
	"github.com/LuiisDev21/stefani/internal/storage"
)

type fakeLLM struct {
	server *httptest.Server
	calls  atomic.Int32
	// handler builds the assistant content for a given request.
	handler func(call int, user string) string
	status  int
}

func newFakeLLM(t *testing.T, handler func(call int, user string) string) *fakeLLM {
	t.Helper()
	f := &fakeLLM{handler: handler, status: http.StatusOK}
	f.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if f.status != http.StatusOK {
			w.WriteHeader(f.status)
			w.Write([]byte(`{"error":{"message":"nope"}}`))
			return
		}
		var body struct {
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		user := ""
		for _, m := range body.Messages {
			if m.Role == "user" {
				user = m.Content
			}
		}
		call := int(f.calls.Add(1))
		content := f.handler(call, user)
		payload, _ := json.Marshal(map[string]any{
			"choices": []any{map[string]any{"message": map[string]any{"content": content}}},
		})
		w.Header().Set("Content-Type", "application/json")
		w.Write(payload)
	}))
	t.Cleanup(f.server.Close)
	return f
}

func (f *fakeLLM) cfg() (config.LLM, string, error) {
	return config.LLM{
		BaseURL:        f.server.URL,
		Model:          "test-model",
		Temperature:    0.2,
		TimeoutSeconds: 5,
		MaxRetries:     1,
		JSONMode:       true,
	}, "test-key", nil
}

func answerFor(lines map[string]string) string {
	items := make([]map[string]any, 0, len(lines))
	for id, text := range lines {
		items = append(items, map[string]any{
			"lineId":        id,
			"language":      "en-US",
			"original":      text,
			"ipa":           "/ˈteɪk ɪt ˈsloʊli/",
			"spanishApprox": "téik it slóuli",
			"translation":   "Tómatelo con calma",
			"notes":         []map[string]string{{"type": "stress", "text": "El acento de 'slowly' esta en la primera silaba."}},
		})
	}
	doc := map[string]any{"lines": items}
	out, _ := json.Marshal(doc)
	return string(out)
}

func setup(t *testing.T, handler func(call int, user string) string) (*Service, *storage.DB, string, *fakeLLM) {
	t.Helper()
	db, err := storage.Open(filepath.Join(t.TempDir(), "ph.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	trackID, err := db.SaveTrack(ctx, domain.Track{
		ID: "local:t1", Provider: domain.ProviderLocal, ProviderID: "t1", Title: "Test",
	})
	if err != nil {
		t.Fatalf("save track: %v", err)
	}
	if err := db.SaveLyrics(ctx, domain.Lyrics{
		TrackID:  trackID,
		Provider: "test",
		Sync:     domain.SyncLine,
		Lines: []domain.Line{
			{ID: "t1-l0", Text: "Take it slowly", StartMs: 0, EndMs: 1000},
			{ID: "t1-l1", Text: "The city hums", StartMs: 1000, EndMs: 2000},
		},
	}); err != nil {
		t.Fatalf("save lyrics: %v", err)
	}

	fake := newFakeLLM(t, handler)
	service := NewService(db, ai.New(5*time.Second), fake.cfg, testLogger{t})
	return service, db, trackID, fake
}

func TestAnnotateTrackStoresValidatedData(t *testing.T) {
	service, db, trackID, fake := setup(t, func(call int, user string) string {
		return answerFor(map[string]string{"t1-l0": "Take it slowly", "t1-l1": "The city hums"})
	})

	var lastProgress Progress
	res, err := service.AnnotateTrack(context.Background(), trackID, "en-US", func(p Progress) { lastProgress = p })
	if err != nil {
		t.Fatalf("annotate: %v", err)
	}
	if res.Generated != 2 || len(res.Failed) != 0 {
		t.Fatalf("result = %+v, want 2 generated", res)
	}
	if lastProgress.Done != 2 || lastProgress.Total != 2 {
		t.Fatalf("progress = %+v", lastProgress)
	}
	annotations, err := db.AnnotationsByTrack(context.Background(), trackID)
	if err != nil {
		t.Fatalf("load annotations: %v", err)
	}
	if len(annotations) != 2 {
		t.Fatalf("stored %d annotations, want 2", len(annotations))
	}
	if annotations[0].IPA != "/ˈteɪk ɪt ˈsloʊli/" {
		t.Fatalf("ipa = %q", annotations[0].IPA)
	}

	// Second run must be served from cache.
	callsBefore := fake.calls.Load()
	res2, err := service.AnnotateTrack(context.Background(), trackID, "en-US", nil)
	if err != nil {
		t.Fatalf("second annotate: %v", err)
	}
	if res2.Generated != 0 || res2.Cached != 2 {
		t.Fatalf("second run = %+v, want everything cached", res2)
	}
	if fake.calls.Load() != callsBefore {
		t.Fatal("cached run called the provider again")
	}
}

func TestAnnotateTrackRejectsInventedLineIDs(t *testing.T) {
	service, db, trackID, _ := setup(t, func(call int, user string) string {
		if call == 1 {
			// Invented id plus missing real lines.
			return answerFor(map[string]string{"ghost-line": "Invisible verse"})
		}
		// Repair attempt: answer only what was really requested.
		return answerFor(map[string]string{"t1-l0": "Take it slowly", "t1-l1": "The city hums"})
	})

	res, err := service.AnnotateTrack(context.Background(), trackID, "en-US", nil)
	if err != nil {
		t.Fatalf("annotate: %v", err)
	}
	if res.Generated != 2 {
		t.Fatalf("generated = %d, want 2 after repair", res.Generated)
	}
	annotations, _ := db.AnnotationsByTrack(context.Background(), trackID)
	for _, a := range annotations {
		if a.LineID == "ghost-line" {
			t.Fatal("an invented line id was stored")
		}
	}
}

func TestAnnotateTrackRejectsChangedOriginal(t *testing.T) {
	service, db, trackID, _ := setup(t, func(call int, user string) string {
		// The model rewrites the verse: that is not the user's lyric.
		return answerFor(map[string]string{"t1-l0": "Take it easy", "t1-l1": "The city hums"})
	})
	res, err := service.AnnotateTrack(context.Background(), trackID, "en-US", nil)
	if err != nil {
		t.Fatalf("annotate: %v", err)
	}
	if res.Generated != 1 {
		t.Fatalf("generated = %d, want 1 (the rewritten line must be rejected)", res.Generated)
	}
	if len(res.Failed) != 1 || res.Failed[0] != "t1-l0" {
		t.Fatalf("failed = %v, want [t1-l0]", res.Failed)
	}
	annotations, _ := db.AnnotationsByTrack(context.Background(), trackID)
	for _, a := range annotations {
		if a.LineID == "t1-l0" {
			t.Fatal("annotation for the rewritten line was stored")
		}
	}
}

func TestAnnotateTrackSurfacesAuthError(t *testing.T) {
	service, _, trackID, fake := setup(t, func(call int, user string) string { return `{}` })
	fake.status = http.StatusUnauthorized
	_, err := service.AnnotateTrack(context.Background(), trackID, "en-US", nil)
	if err == nil {
		t.Fatal("expected an error for a rejected key")
	}
	if !strings.Contains(err.Error(), "credenciales") && !strings.Contains(err.Error(), "401") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNormalizeIPA(t *testing.T) {
	cases := map[string]string{
		"teɪk ɪt":        "/teɪk ɪt/",
		"/teɪk ɪt/":      "/teɪk ɪt/",
		"take it slowly": "",
		"":               "",
	}
	for in, want := range cases {
		if got := normalizeIPA(in); got != want {
			t.Fatalf("normalizeIPA(%q) = %q, want %q", in, got, want)
		}
	}
}

type testLogger struct{ t *testing.T }

func (l testLogger) Info(msg string, args ...any) { l.t.Logf(msg, args...) }
func (l testLogger) Warn(msg string, args ...any) { l.t.Logf("WARN: "+fmt.Sprint(msg), args...) }
