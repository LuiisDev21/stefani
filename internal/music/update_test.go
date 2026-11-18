package music

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdateInstallsTheReleaseAsset(t *testing.T) {
	payload := strings.Repeat("x", 600<<10)
	mux := http.NewServeMux()
	mux.HandleFunc("/latest/download/"+assetName(), func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/download/2026.11.01/"+assetName(), http.StatusFound)
	})
	mux.HandleFunc("/download/2026.11.01/"+assetName(), func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(payload))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	oldBase := releaseBase
	releaseBase = server.URL + "/latest/download/"
	defer func() { releaseBase = oldBase }()

	dir := t.TempDir()
	oldDir := managedDir
	SetManagedDir(dir)
	defer SetManagedDir(oldDir)

	client := &YtDlp{}
	version, err := client.Update(context.Background())
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if version != "2026.11.01" {
		t.Fatalf("unexpected version %q", version)
	}
	data, err := os.ReadFile(filepath.Join(dir, "yt-dlp"+exeSuffix()))
	if err != nil {
		t.Fatalf("read downloaded binary: %v", err)
	}
	if len(data) != len(payload) {
		t.Fatalf("downloaded %d bytes, want %d", len(data), len(payload))
	}
}

func TestUpdateRejectsShortDownloads(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/latest/download/"+assetName(), func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/download/1/"+assetName(), http.StatusFound)
	})
	mux.HandleFunc("/download/1/"+assetName(), func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("tiny"))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	oldBase := releaseBase
	releaseBase = server.URL + "/latest/download/"
	defer func() { releaseBase = oldBase }()

	dir := t.TempDir()
	oldDir := managedDir
	SetManagedDir(dir)
	defer SetManagedDir(oldDir)

	if _, err := (&YtDlp{}).Update(context.Background()); err == nil {
		t.Fatal("expected an error for a truncated download")
	}
	if _, err := os.Stat(filepath.Join(dir, "yt-dlp"+exeSuffix())); !os.IsNotExist(err) {
		t.Fatal("a failed download must not replace the binary")
	}
}

func exeSuffix() string {
	if strings.HasSuffix(assetName(), ".exe") {
		return ".exe"
	}
	return ""
}
