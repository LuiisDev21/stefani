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
	server, err := NewServer(func(r *http.Request, trackID string) (Source, error) {
		if trackID == "ok" {
			return Source{Path: audio}, nil
		}
		return Source{}, ErrNoAudio
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
	resp, err := http.Get(server.URL("..%2F..%2Fetc%2Fpasswd"))
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		t.Fatalf("traversal was served: %d", resp.StatusCode)
	}
}

func TestStreamHandlerResolvesLocalSource(t *testing.T) {
	audio := filepath.Join(t.TempDir(), "x.wav")
	writeTestWAV(t, audio, 0.5)
	handler := StreamHandler(func(r *http.Request, trackID string) (Source, error) {
		if trackID == "x" {
			return Source{Path: audio}, nil
		}
		return Source{}, os.ErrNotExist
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

// proxyFixture points the relay at a local TLS server and restores the real
// host rule and client afterwards.
func proxyFixture(t *testing.T, upstream *httptest.Server, seen *string) http.Handler {
	t.Helper()
	origHost, origClient := allowedRemoteHost, relayClient
	allowedRemoteHost = func(host string) bool { return host == "127.0.0.1" }
	relayClient = upstream.Client()
	t.Cleanup(func() { allowedRemoteHost, relayClient = origHost, origClient })

	return StreamHandler(func(r *http.Request, trackID string) (Source, error) {
		return Source{URL: upstream.URL + "/videoplayback?itag=251"}, nil
	})
}

func TestProxyRelaysRangeRequests(t *testing.T) {
	var seenRange string
	upstream := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenRange = r.Header.Get("Range")
		body := make([]byte, 1024)
		for i := range body {
			body[i] = 0x1a
		}
		w.Header().Set("Content-Type", "audio/webm")
		w.Header().Set("Content-Range", "bytes 0-1023/4605080")
		w.Header().Set("Accept-Ranges", "bytes")
		w.WriteHeader(http.StatusPartialContent)
		w.Write(body)
	}))
	defer upstream.Close()

	srv := httptest.NewServer(proxyFixture(t, upstream, &seenRange))
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/media/yt", nil)
	req.Header.Set("Range", "bytes=0-1023")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("proxy request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent {
		t.Fatalf("status = %d, want 206", resp.StatusCode)
	}
	if seenRange != "bytes=0-1023" {
		t.Fatalf("upstream saw range %q", seenRange)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "audio/webm" {
		t.Fatalf("content type = %q", ct)
	}
	if cr := resp.Header.Get("Content-Range"); cr != "bytes 0-1023/4605080" {
		t.Fatalf("content-range = %q", cr)
	}
	body, _ := io.ReadAll(resp.Body)
	if len(body) != 1024 {
		t.Fatalf("relayed body = %d bytes, want 1024", len(body))
	}
	if resp.Header.Get("Cache-Control") != "private, no-store" {
		t.Fatalf("cache-control = %q", resp.Header.Get("Cache-Control"))
	}
}

func TestProxyRejectsForeignHosts(t *testing.T) {
	handler := StreamHandler(func(r *http.Request, trackID string) (Source, error) {
		return Source{URL: "https://evil.example/audio.m4a"}, nil
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
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestProxyRejectsPlainHTTP(t *testing.T) {
	handler := StreamHandler(func(r *http.Request, trackID string) (Source, error) {
		return Source{URL: "http://rr1---sn-x.googlevideo.com/audio"}, nil
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
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestProxyRejectsEmptySource(t *testing.T) {
	handler := StreamHandler(func(r *http.Request, trackID string) (Source, error) {
		return Source{}, nil
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
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}
