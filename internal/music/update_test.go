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
	mux.HandleFunc("/release", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tag_name":"2026.11.01","assets":[{"name":"` + assetName() + `","browser_download_url":"` +
			"http://" + r.Host + "/binary" + `"}]}`))
	})
	mux.HandleFunc("/binary", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(payload))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	oldEndpoint := releaseEndpoint
	releaseEndpoint = server.URL + "/release"
	defer func() { releaseEndpoint = oldEndpoint }()

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
	mux.HandleFunc("/release", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"1","assets":[{"name":"` + assetName() + `","browser_download_url":"` +
			"http://" + r.Host + "/binary" + `"}]}`))
	})
	mux.HandleFunc("/binary", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("tiny"))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	oldEndpoint := releaseEndpoint
	releaseEndpoint = server.URL + "/release"
	defer func() { releaseEndpoint = oldEndpoint }()

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
