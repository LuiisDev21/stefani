package media

import (
	"encoding/binary"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTestWAV creates a short valid WAV so ServeContent has a real file.
func writeTestWAV(t *testing.T, path string, seconds float64) {
	t.Helper()
	spec := Spec{
		SampleRate: 8000,
		Duration:   seconds,
		Notes:      []Note{{StartSec: 0, DurSec: seconds, Freq: 440, Amp: 0.2, Timbre: TimbreMelody}},
	}
	if err := os.WriteFile(path, RenderWAV(spec), 0o644); err != nil {
		t.Fatalf("write wav: %v", err)
	}
}

func newTestServer(t *testing.T) (*Server, string) {
	t.Helper()
	dir := t.TempDir()
	audio := filepath.Join(dir, "track.wav")
	writeTestWAV(t, audio, 1)
	server, err := NewServer(func(r *http.Request, trackID string) (string, error) {
		if trackID == "ok" {
			return audio, nil
		}
		return "", ErrNoAudio
	})
	if err != nil {
		t.Fatalf("new server: %v", err)
	}
	t.Cleanup(func() { server.Close() })
	return server, audio
}

func TestServerServesFullFile(t *testing.T) {
	server, audio := newTestServer(t)
	resp, err := http.Get(server.URL("ok"))
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	info, _ := os.Stat(audio)
	if int64(len(body)) != info.Size() {
		t.Fatalf("body = %d bytes, want %d", len(body), info.Size())
	}
	if string(body[0:4]) != "RIFF" {
		t.Fatalf("not a wav: %q", body[0:4])
	}
	if resp.Header.Get("Accept-Ranges") != "bytes" {
		t.Fatal("missing Accept-Ranges")
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "audio/wav") {
		t.Fatalf("content type = %q", ct)
	}
}

func TestServerSupportsRangeRequests(t *testing.T) {
	server, _ := newTestServer(t)
	req, _ := http.NewRequest(http.MethodGet, server.URL("ok"), nil)
	req.Header.Set("Range", "bytes=0-1023")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("range request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusPartialContent {
		t.Fatalf("status = %d, want 206", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if len(body) != 1024 {
		t.Fatalf("body = %d bytes, want 1024", len(body))
	}
	if cr := resp.Header.Get("Content-Range"); !strings.HasPrefix(cr, "bytes 0-1023/") {
		t.Fatalf("content-range = %q", cr)
	}
	if binary.LittleEndian.Uint32(body[24:28]) != 8000 {
		t.Fatalf("sample rate not present in the ranged body")
	}
}

func TestServerRequiresToken(t *testing.T) {
	server, _ := newTestServer(t)
	resp, err := http.Get(server.BaseURL() + "/media/ok")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", resp.StatusCode)
	}
}

func TestServerRejectsUnknownTrack(t *testing.T) {
	server, _ := newTestServer(t)
	resp, err := http.Get(server.URL("missing"))
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestServerRejectsTraversal(t *testing.T) {
	server, _ := newTestServer(t)
	// The token is valid; the path is not.
	resp, err := http.Get(server.URL("..%2F..%2Fetc%2Fpasswd"))
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		t.Fatalf("traversal was served: %d", resp.StatusCode)
	}
}

func TestStreamHandlerResolvesThroughMux(t *testing.T) {
	// The asset-server middleware path uses the same handler shape.
	audio := filepath.Join(t.TempDir(), "x.wav")
	writeTestWAV(t, audio, 0.5)
	handler := StreamHandler(func(r *http.Request, trackID string) (string, error) {
		if trackID == "x" {
			return audio, nil
		}
		return "", os.ErrNotExist
	})
	mux := http.NewServeMux()
	mux.Handle("/media/", handler)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/media/x")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if resp.Header.Get("Content-Type") != "audio/wav" {
		t.Fatalf("content type = %q", resp.Header.Get("Content-Type"))
	}
}
