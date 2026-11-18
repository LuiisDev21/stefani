package music

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
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
)

// Runner executes an external command. It is injectable so tests never shell
// out and CI does not need yt-dlp.
type Runner func(ctx context.Context, name string, args ...string) (stdout, stderr []byte, err error)

func execRunner(ctx context.Context, name string, args ...string) ([]byte, []byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
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
	version string
	run     Runner
	now     func() time.Time
	ttl     time.Duration

	mu    sync.Mutex
	cache map[string]cachedStream
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
	ctx, cancel := context.WithTimeout(context.Background(), probeTimeout)
	defer cancel()
	if out, _, err := y.run(ctx, path, "--version"); err == nil {
		if v := strings.TrimSpace(string(out)); v != "" && len(v) < 40 {
			y.version = v
		}
	}
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
	return y.version
}

// Detail is a short human-readable status for the settings panel.
func (y *YtDlp) Detail() string {
	if y.Available() {
		if y.version != "" {
			return "yt-dlp " + y.version
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

// sanitizeStderr keeps the useful part of an error message without leaking
// URLs or cookies into logs.
func sanitizeStderr(stderr []byte) string {
	line := strings.TrimSpace(string(stderr))
	if idx := strings.IndexByte(line, '\n'); idx > 0 {
		line = line[:idx]
	}
	line = strings.ReplaceAll(line, "https://", "hxxps://")
	if len(line) > 180 {
		line = line[:180]
	}
	return line
}
