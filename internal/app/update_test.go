package app

import (
	"archive/zip"
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestIsNewer(t *testing.T) {
	cases := []struct {
		latest, current string
		want            bool
	}{
		{"v0.1.10", "v0.1.9", true},
		{"v0.2.0", "v0.1.99", true},
		{"v0.1.10", "v0.1.10", false},
		{"v0.1.9", "v0.1.10", false},
		{"v0.1.10", "dev", false},
		{"basura", "v0.1.0", false},
	}
	for _, c := range cases {
		if got := isNewer(c.latest, c.current); got != c.want {
			t.Errorf("isNewer(%q, %q) = %v, want %v", c.latest, c.current, got, c.want)
		}
	}
}

func TestLatestTagFollowsRedirect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/releases/tag/v9.9.9", http.StatusFound)
	}))
	defer server.Close()

	old := releasesLatestURL
	releasesLatestURL = server.URL
	defer func() { releasesLatestURL = old }()

	tag, err := latestTag(updateClient)
	if err != nil {
		t.Fatalf("latestTag: %v", err)
	}
	if tag != "v9.9.9" {
		t.Fatalf("tag = %q", tag)
	}
}

func TestExtractAsset(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, _ := zw.Create(installerName)
	_, _ = w.Write([]byte("installer-bytes"))
	_ = zw.Close()

	target := filepath.Join(t.TempDir(), installerName)
	if err := extractAsset(buf.Bytes(), installerName, target); err != nil {
		t.Fatalf("extractAsset: %v", err)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(data) != "installer-bytes" {
		t.Fatalf("content = %q", data)
	}
}

func TestExtractAssetMissing(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	_, _ = zw.Create("otro.exe")
	_ = zw.Close()
	if err := extractAsset(buf.Bytes(), installerName, filepath.Join(t.TempDir(), installerName)); err == nil {
		t.Fatal("expected an error when the installer is absent")
	}
}
