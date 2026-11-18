package music

import (
	"os"
	"path/filepath"
	"testing"
)

// The app must find a yt-dlp installed in Python's Scripts folder even when
// that folder is not on PATH (the usual case when launching from Explorer).
func TestFindYtDlpUsesKnownWindowsLocations(t *testing.T) {
	oldManaged := managedDir
	SetManagedDir("")
	defer SetManagedDir(oldManaged)

	root := t.TempDir()
	scripts := filepath.Join(root, "Programs", "Python", "Python312", "Scripts")
	if err := os.MkdirAll(scripts, 0o755); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(scripts, "yt-dlp.exe")
	if err := os.WriteFile(binary, []byte("stub"), 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PATH", t.TempDir())
	t.Setenv("LOCALAPPDATA", root)
	t.Setenv("APPDATA", t.TempDir())
	t.Setenv("USERPROFILE", t.TempDir())
	t.Setenv("ProgramData", t.TempDir())

	got := findYtDlp()
	want, err := filepath.Abs(binary)
	if err != nil {
		t.Fatal(err)
	}
	gotAbs, err := filepath.Abs(got)
	if err != nil {
		t.Fatal(err)
	}
	if gotAbs != want {
		t.Fatalf("findYtDlp() = %q, want %q", got, binary)
	}
}
