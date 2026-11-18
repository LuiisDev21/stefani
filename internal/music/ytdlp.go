package music

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/LuiisDev21/stefani/internal/executil"
)

// Online playback note
//
// STEFANI never touches YouTube's audio by itself. When the user enables online
// playback, a locally installed yt-dlp resolves the audio URL and this package
// only validates and caches it. The binary is not bundled: the app detects it,
// reports whether it is present, and the UI explains the requirement.
//
// The resolved URL points at googlevideo and expires (usually in a few hours).
// It is never logged, never persisted and never sent to the frontend: the media
// proxy in internal/media relays the bytes.

var (
	// ErrYtDlpMissing means online playback was requested but the helper is not
	// installed on this machine.
	ErrYtDlpMissing = errors.New("yt-dlp is not installed")
	// ErrStreamResolve means yt-dlp ran but did not return a usable URL.
	ErrStreamResolve = errors.New("could not resolve the audio stream")
)

var videoIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{11}$`)

const (
	resolveTimeout   = 25 * time.Second
	probeTimeout     = 5 * time.Second
	urlSafetyMargin  = 90 * time.Second
	defaultStreamTTL = 4 * time.Hour
	maxCacheEntries  = 256
	warmTimeout      = 12 * time.Second
)

// Runner executes an external command. It is injectable so tests never shell
// out and CI does not need yt-dlp.
type Runner func(ctx context.Context, name string, args ...string) (stdout, stderr []byte, err error)

func execRunner(ctx context.Context, name string, args ...string) ([]byte, []byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	executil.Hide(cmd)
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	err := cmd.Run()
	return out.Bytes(), errBuf.Bytes(), err
}

// StreamURL is a resolved, time-limited audio URL.
type StreamURL struct {
	URL       string
	ExpiresAt time.Time
	VideoID   string
}

type cachedStream struct {
	url       string
	expiresAt time.Time
}

// YtDlp resolves audio URLs through a local yt-dlp binary.
type YtDlp struct {
	path    string
	run     Runner
	now     func() time.Time
	ttl     time.Duration

	verMu   sync.RWMutex
	version string

	mu       sync.Mutex
	cache    map[string]cachedStream
	inflight map[string]chan struct{}
}

// NewYtDlpWithRunner builds a resolver with an injected runner, for tests and
// for wiring a different binary location.
func NewYtDlpWithRunner(path string, run Runner, now func() time.Time, ttl time.Duration) *YtDlp {
	if now == nil {
		now = time.Now
	}
	if ttl <= 0 {
		ttl = defaultStreamTTL
	}
	return &YtDlp{
		path:    path,
		version: "test",
		run:     run,
		now:     now,
		ttl:     ttl,
		cache:   make(map[string]cachedStream),
	}
}

// managedDir is where the app keeps its own yt-dlp copy. It is set from main
// so the file follows STEFANI_HOME.
var managedDir string

// SetManagedDir points the updater at the folder that holds the app's copy.
func SetManagedDir(dir string) { managedDir = dir }

// managedBinary is the path of the copy the app owns and updates.
func managedBinary() string {
	dir := managedDir
	if dir == "" {
		base, err := os.UserConfigDir()
		if err != nil {
			return ""
		}
		dir = filepath.Join(base, "stefani", "bin")
	}
	name := "yt-dlp"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return filepath.Join(dir, name)
}

// findYtDlp returns the binary to run: the app's own copy first (so updates
// win), then PATH and the usual Windows install locations. Without the latter,
// launching the app from Explorer misses a yt-dlp whose folder is not on the
// user's PATH.
func findYtDlp() string {
	if path := managedBinary(); path != "" {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path
		}
	}
	if path, err := exec.LookPath("yt-dlp"); err == nil {
		return path
	}
	var candidates []string
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		candidates = append(candidates, filepath.Join(dir, "yt-dlp.exe"), filepath.Join(dir, "yt-dlp"))
	}
	if local := os.Getenv("LOCALAPPDATA"); local != "" {
		matches, _ := filepath.Glob(filepath.Join(local, "Programs", "Python", "Python3*", "Scripts", "yt-dlp.exe"))
		candidates = append(candidates, matches...)
		candidates = append(candidates, filepath.Join(local, "Microsoft", "WinGet", "Links", "yt-dlp.exe"))
	}
	if roaming := os.Getenv("APPDATA"); roaming != "" {
		matches, _ := filepath.Glob(filepath.Join(roaming, "Python", "Python3*", "Scripts", "yt-dlp.exe"))
		candidates = append(candidates, matches...)
	}
	if profile := os.Getenv("USERPROFILE"); profile != "" {
		candidates = append(candidates, filepath.Join(profile, "scoop", "shims", "yt-dlp.exe"))
	}
	if programData := os.Getenv("ProgramData"); programData != "" {
		candidates = append(candidates, filepath.Join(programData, "chocolatey", "bin", "yt-dlp.exe"))
	}
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate
		}
	}
	return ""
}

// DetectYtDlp looks for the binary and records its version. A missing binary is
// a normal state, not an error: the caller reports it to the UI.
func DetectYtDlp() *YtDlp {
	y := &YtDlp{
		run:   execRunner,
		now:   time.Now,
		ttl:   defaultStreamTTL,
		cache: make(map[string]cachedStream),
	}
	path := findYtDlp()
	if path == "" {
		return y
	}
	y.path = path
	// The version probe costs seconds on Windows (a packed Python binary), so it
	// runs in the background: the window opens before it answers.
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), probeTimeout)
		defer cancel()
		if out, _, err := y.run(ctx, path, "--version"); err == nil {
			if v := strings.TrimSpace(string(out)); v != "" && len(v) < 40 {
				y.verMu.Lock()
				y.version = v
				y.verMu.Unlock()
			}
		}
	}()
	y.loadCache()
	return y
}

// Available reports whether a usable yt-dlp was found.
func (y *YtDlp) Available() bool {
	return y != nil && y.path != ""
}

// Version is the detected yt-dlp version, empty when unavailable.
func (y *YtDlp) Version() string {
	if y == nil {
		return ""
	}
	y.verMu.RLock()
	defer y.verMu.RUnlock()
	return y.version
}

// Detail is a short human-readable status for the settings panel.
func (y *YtDlp) Detail() string {
	if y.Available() {
		if v := y.Version(); v != "" {
			return "yt-dlp " + v
		}
		return "yt-dlp detectado"
	}
	return "yt-dlp no encontrado"
}

// Resolve returns a direct audio URL for a YouTube video id, using the cache
// while the URL is still valid.
func (y *YtDlp) Resolve(ctx context.Context, videoID string) (StreamURL, error) {
	if !y.Available() {
		return StreamURL{}, ErrYtDlpMissing
	}
	videoID = strings.TrimSpace(videoID)
	if !videoIDPattern.MatchString(videoID) {
		return StreamURL{}, fmt.Errorf("%w: identificador invalido", ErrStreamResolve)
	}
	if cached, ok := y.lookup(videoID); ok {
		return cached, nil
	}
	// A warm-up and a play can ask for the same song at the same time; one
	// process answers both.
	if wait, ok := y.claim(videoID); ok {
		select {
		case <-wait:
		case <-ctx.Done():
			return StreamURL{}, ctx.Err()
		}
		if cached, ok := y.lookup(videoID); ok {
			return cached, nil
		}
		return StreamURL{}, fmt.Errorf("%w: no se pudo resolver", ErrStreamResolve)
	}
	defer y.release(videoID)

	runCtx, cancel := context.WithTimeout(ctx, resolveTimeout)
	defer cancel()
	args := []string{
		"-f", "bestaudio",
		"-g",
		"--no-playlist",
		"--no-warnings",
		"--no-progress",
		"--no-call-home",
		"https://music.youtube.com/watch?v=" + videoID,
	}
	stdout, stderr, err := y.run(runCtx, y.path, args...)
	if err != nil {
		if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
			return StreamURL{}, fmt.Errorf("%w: yt-dlp agoto el tiempo de espera", ErrStreamResolve)
		}
		detail := sanitizeStderr(stderr)
		if detail == "" {
			detail = "yt-dlp no devolvio una URL"
		}
		return StreamURL{}, fmt.Errorf("%w: %s", ErrStreamResolve, detail)
	}

	raw := firstURL(stdout)
	if raw == "" {
		return StreamURL{}, fmt.Errorf("%w: respuesta vacia de yt-dlp", ErrStreamResolve)
	}
	expiresAt, err := validateStreamURL(raw)
	if err != nil {
		return StreamURL{}, err
	}
	if expiresAt.IsZero() {
		expiresAt = y.now().Add(y.ttl)
	}
	out := StreamURL{URL: raw, ExpiresAt: expiresAt, VideoID: videoID}
	y.store(videoID, out)
	return out, nil
}

// IsAllowedStreamHost limits URLs to Google's media hosts. It is exported so the
// media proxy can enforce the same rule at the byte-relay boundary.
func IsAllowedStreamHost(host string) bool {
	host = strings.ToLower(strings.TrimSuffix(host, "."))
	if host == "" {
		return false
	}
	return host == "googlevideo.com" || strings.HasSuffix(host, ".googlevideo.com")
}

func validateStreamURL(raw string) (time.Time, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: URL malformada", ErrStreamResolve)
	}
	if parsed.Scheme != "https" {
		return time.Time{}, fmt.Errorf("%w: la URL no es https", ErrStreamResolve)
	}
	if !IsAllowedStreamHost(parsed.Hostname()) {
		return time.Time{}, fmt.Errorf("%w: host no permitido", ErrStreamResolve)
	}
	if seconds, err := strconv.ParseInt(parsed.Query().Get("expire"), 10, 64); err == nil && seconds > 0 {
		return time.Unix(seconds, 0).Add(-urlSafetyMargin), nil
	}
	return time.Time{}, nil
}

func (y *YtDlp) lookup(videoID string) (StreamURL, bool) {
	y.mu.Lock()
	defer y.mu.Unlock()
	entry, ok := y.cache[videoID]
	if !ok || !y.now().Before(entry.expiresAt) {
		delete(y.cache, videoID)
		return StreamURL{}, false
	}
	return StreamURL{URL: entry.url, ExpiresAt: entry.expiresAt, VideoID: videoID}, true
}

func (y *YtDlp) store(videoID string, s StreamURL) {
	y.mu.Lock()
	defer y.mu.Unlock()
	if len(y.cache) >= maxCacheEntries {
		for id, entry := range y.cache {
			if !y.now().Before(entry.expiresAt) {
				delete(y.cache, id)
			}
		}
		if len(y.cache) >= maxCacheEntries {
			y.cache = make(map[string]cachedStream, maxCacheEntries)
		}
	}
	y.cache[videoID] = cachedStream{url: s.URL, expiresAt: s.ExpiresAt}
	go y.saveCache()
}

// firstURL picks the first http(s) line of yt-dlp's stdout.
func firstURL(stdout []byte) string {
	for _, line := range strings.Split(string(stdout), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "https://") {
			return line
		}
	}
	return ""
}

// sanitizeStderr keeps the last line, which is the one that says what went
// wrong, and drops the ERROR: prefix the UI does not need.
func sanitizeStderr(stderr []byte) string {
	text := strings.TrimSpace(string(stderr))
	if text == "" {
		return ""
	}
	lines := strings.Split(text, "\n")
	last := strings.TrimSpace(lines[len(lines)-1])
	last = strings.TrimPrefix(last, "ERROR: ")
	// Signed media URLs carry tokens; the UI never needs them, so the scheme is
	// defused rather than dropped and the line still explains the failure.
	last = strings.ReplaceAll(last, "https://", "hxxps://")
	last = strings.ReplaceAll(last, "http://", "hxxp://")
	if len(last) > 240 {
		last = last[:240] + "…"
	}
	return last
}


// SetCacheFile points the resolved-URL cache at a file so a URL found before a
// restart is still usable: they live for hours, the process does not.
func SetCacheFile(path string) {
	cacheMu.Lock()
	cacheFile = path
	cacheMu.Unlock()
}

var (
	cacheMu   sync.Mutex
	cacheFile string
)

// loadCache seeds the in-memory cache from disk, dropping anything expired.
func (y *YtDlp) loadCache() {
	cacheMu.Lock()
	path := cacheFile
	cacheMu.Unlock()
	if path == "" {
		return
	}
	blob, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var saved map[string]struct {
		URL       string    `json:"url"`
		ExpiresAt time.Time `json:"expiresAt"`
	}
	if json.Unmarshal(blob, &saved) != nil {
		return
	}
	now := y.now()
	y.mu.Lock()
	defer y.mu.Unlock()
	for id, entry := range saved {
		if entry.URL != "" && now.Before(entry.ExpiresAt) {
			y.cache[id] = cachedStream{url: entry.URL, expiresAt: entry.ExpiresAt}
		}
	}
}

// saveCache writes the live entries back. It runs off the request path.
func (y *YtDlp) saveCache() {
	cacheMu.Lock()
	path := cacheFile
	cacheMu.Unlock()
	if path == "" {
		return
	}
	now := y.now()
	y.mu.Lock()
	saved := make(map[string]struct {
		URL       string    `json:"url"`
		ExpiresAt time.Time `json:"expiresAt"`
	}, len(y.cache))
	for id, entry := range y.cache {
		if now.Before(entry.expiresAt) {
			saved[id] = struct {
				URL       string    `json:"url"`
				ExpiresAt time.Time `json:"expiresAt"`
			}{URL: entry.url, ExpiresAt: entry.expiresAt}
		}
	}
	y.mu.Unlock()
	blob, err := json.Marshal(saved)
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	_ = os.WriteFile(path, blob, 0o600)
}

// claim registers a resolve in flight. It returns a channel to wait on when
// another caller is already resolving the same video, and ok=false when this
// caller owns the work.
func (y *YtDlp) claim(videoID string) (chan struct{}, bool) {
	y.mu.Lock()
	defer y.mu.Unlock()
	if y.inflight == nil {
		y.inflight = make(map[string]chan struct{})
	}
	if wait, busy := y.inflight[videoID]; busy {
		return wait, true
	}
	y.inflight[videoID] = make(chan struct{})
	return nil, false
}

func (y *YtDlp) release(videoID string) {
	y.mu.Lock()
	defer y.mu.Unlock()
	if wait, ok := y.inflight[videoID]; ok {
		close(wait)
		delete(y.inflight, videoID)
	}
}
