package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadMigratesLegacyConfigToOnlinePlayback(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	legacy := `{"llm":{"baseUrl":"https://api.openai.com/v1","model":"gpt-4o-mini"},"features":{"youtubeMusic":false},"accent":"en-US"}`
	if err := os.WriteFile(path, []byte(legacy), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	cfg, err := NewStore(path).Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if !cfg.Features.YouTubeMusic {
		t.Fatal("a document without schema must get the new default: online playback on")
	}
	if cfg.Schema != CurrentSchema {
		t.Fatalf("schema = %d, want %d", cfg.Schema, CurrentSchema)
	}
}

func TestLoadRespectsTheUserChoiceAfterMigration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	current := `{"schema":2,"llm":{"baseUrl":"https://api.openai.com/v1","model":"gpt-4o-mini"},"features":{"youtubeMusic":false}}`
	if err := os.WriteFile(path, []byte(current), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	cfg, err := NewStore(path).Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Features.YouTubeMusic {
		t.Fatal("a current document must keep the flag the user turned off")
	}
}

func TestDefaultEnablesOnlinePlayback(t *testing.T) {
	if !Default().Features.YouTubeMusic {
		t.Fatal("online playback is the main source of music and ships enabled")
	}
}

func TestClientIDIsStableAcrossLoads(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(filepath.Join(dir, "config.json"))
	first, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	second, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if first.LLM.ClientID == "" {
		t.Fatal("client id must be generated on load")
	}
	if first.LLM.ClientID != second.LLM.ClientID {
		t.Fatalf("client id changed: %q -> %q", first.LLM.ClientID, second.LLM.ClientID)
	}
	if strings.Contains(readFile(t, filepath.Join(dir, "config.json")), first.LLM.ClientID) {
		t.Fatal("client id must not be written into config.json")
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}
