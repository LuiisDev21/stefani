package config

import (
	"os"
	"path/filepath"
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
