package music

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// releaseDownloadBase points at GitHub's "latest" redirect, which resolves the
// asset without the API and therefore without its 60-requests-per-hour limit.
const releaseDownloadBase = "https://github.com/yt-dlp/yt-dlp/releases/latest/download/"

// releaseBase is a variable so tests can point at a fake GitHub.
var releaseBase = releaseDownloadBase

// ErrReleaseUnavailable means GitHub did not answer with a usable release.
var ErrReleaseUnavailable = errors.New("no se pudo consultar la última versión de yt-dlp")

// assetName is the release asset for this platform. yt-dlp publishes a
// standalone binary per OS, so no Python is needed on the other machine.
func assetName() string {
	switch runtime.GOOS {
	case "windows":
		return "yt-dlp.exe"
	case "darwin":
		return "yt-dlp_macos"
	default:
		if runtime.GOARCH == "arm64" {
			return "yt-dlp_linux_aarch64"
		}
		return "yt-dlp_linux"
	}
}

// Update downloads the latest yt-dlp release into the app's bin folder and
// swaps it in atomically. It returns the new version tag.
func (y *YtDlp) Update(ctx context.Context) (string, error) {
	target := managedBinary()
	if target == "" {
		return "", errors.New("no hay una carpeta para guardar yt-dlp")
	}
	asset := assetName()
	source := releaseBase + asset
	version := ""
	// The tag is in the first redirect; reading it avoids downloading the
	// binary just to learn which version it is.
	if tag, location, err := resolve(ctx, source); err == nil {
		version = tag
		source = location
	}
	if _, err := download(ctx, source, target); err != nil {
		return "", err
	}
	return version, nil
}

// resolve follows the "latest" redirect only once to read the tag.
func resolve(ctx context.Context, raw string) (string, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", "stefani")
	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	res, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()
	location := res.Header.Get("Location")
	if res.StatusCode < 300 || res.StatusCode > 399 || location == "" {
		return "", raw, nil
	}
	if base, err := url.Parse(raw); err == nil {
		if ref, err := base.Parse(location); err == nil {
			location = ref.String()
		}
	}
	return versionFromURL(location), location, nil
}

// download writes to a temporary file next to the target and renames it, so a
// failed half-download never replaces a working binary. The version comes from
// the redirect target (.../releases/download/<tag>/<asset>).
func download(ctx context.Context, url, target string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "stefani")
	client := &http.Client{Timeout: 5 * time.Minute}
	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrReleaseUnavailable, err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: HTTP %d", ErrReleaseUnavailable, res.StatusCode)
	}

	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return "", err
	}
	tmp, err := os.CreateTemp(filepath.Dir(target), "yt-dlp-*")
	if err != nil {
		return "", err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	size, err := io.Copy(tmp, io.LimitReader(res.Body, 200<<20))
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return "", fmt.Errorf("descargar yt-dlp: %w", err)
	}
	if size < 500<<10 {
		return "", errors.New("la descarga de yt-dlp quedó incompleta")
	}
	if err := os.Chmod(tmpName, 0o755); err != nil {
		return "", err
	}
	if err := os.Rename(tmpName, target); err != nil {
		return "", fmt.Errorf("reemplazar yt-dlp: %w", err)
	}
	return versionFromURL(res.Request.URL.Path), nil
}

// versionFromURL reads the tag out of .../releases/download/<tag>/<asset>.
func versionFromURL(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) < 2 {
		return ""
	}
	return parts[len(parts)-2]
}
