package music

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const releaseAPI = "https://api.github.com/repos/yt-dlp/yt-dlp/releases/latest"

// releaseEndpoint is a variable so tests can point at a fake GitHub.
var releaseEndpoint = releaseAPI

type release struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
}

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

	info, err := fetchRelease(ctx)
	if err != nil {
		return "", err
	}
	asset := assetName()
	var url string
	for _, item := range info.Assets {
		if item.Name == asset {
			url = item.URL
			break
		}
	}
	if url == "" {
		return "", fmt.Errorf("%w: no hay binario para %s/%s", ErrReleaseUnavailable, runtime.GOOS, runtime.GOARCH)
	}

	if err := download(ctx, url, target); err != nil {
		return "", err
	}
	return info.TagName, nil
}

func fetchRelease(ctx context.Context) (release, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, releaseEndpoint, nil)
	if err != nil {
		return release{}, err
	}
	req.Header.Set("User-Agent", "stefani")
	req.Header.Set("Accept", "application/vnd.github+json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return release{}, fmt.Errorf("%w: %v", ErrReleaseUnavailable, err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return release{}, fmt.Errorf("%w: HTTP %d", ErrReleaseUnavailable, res.StatusCode)
	}
	var info release
	if err := json.NewDecoder(io.LimitReader(res.Body, 2<<20)).Decode(&info); err != nil {
		return release{}, fmt.Errorf("%w: respuesta ilegible", ErrReleaseUnavailable)
	}
	if info.TagName == "" {
		return release{}, ErrReleaseUnavailable
	}
	return info, nil
}

// download writes to a temporary file next to the target and renames it, so a
// failed half-download never replaces a working binary.
func download(ctx context.Context, url, target string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "stefani")
	client := &http.Client{Timeout: 5 * time.Minute}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("descargar yt-dlp: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("descargar yt-dlp: HTTP %d", res.StatusCode)
	}

	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(target), "yt-dlp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	size, err := io.Copy(tmp, io.LimitReader(res.Body, 200<<20))
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return fmt.Errorf("descargar yt-dlp: %w", err)
	}
	if size < 500<<10 {
		return errors.New("la descarga de yt-dlp quedó incompleta")
	}
	if err := os.Chmod(tmpName, 0o755); err != nil {
		return err
	}
	if err := os.Rename(tmpName, target); err != nil {
		return fmt.Errorf("reemplazar yt-dlp: %w", err)
	}
	return nil
}
