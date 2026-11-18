package app

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/LuiisDev21/stefani/internal/brand"
	"github.com/wailsapp/wails/v3/pkg/application"
)

const (
	releasesLatest   = "https://github.com/LuiisDev21/stefani/releases/latest"
	releasesDownload = "https://github.com/LuiisDev21/stefani/releases/download/"
	windowsAsset     = "stefani-windows-amd64.zip"
	installerName    = "stefani-amd64-installer.exe"
)

// updateClient and releasesLatestURL are variables so tests can point at a
// fake GitHub instead of the network.
var (
	updateClient      = &http.Client{Timeout: 5 * time.Minute}
	releasesLatestURL = releasesLatest
)

// CheckAppUpdate compares the running build with the latest published tag.
// It never fails: the panel shows "no se pudo comprobar" instead.
func (a *App) CheckAppUpdate() AppUpdateDTO {
	dto := AppUpdateDTO{Current: brand.Version, URL: releasesLatest}
	tag, err := latestTag(updateClient)
	if err != nil {
		dto.Detail = "No se pudo comprobar la última versión"
		return dto
	}
	dto.Latest = tag
	dto.Available = isNewer(tag, brand.Version)
	if dto.Available {
		dto.Detail = "Hay una versión nueva"
	} else {
		dto.Detail = "Estás en la última versión"
	}
	return dto
}

// InstallAppUpdate downloads the platform archive, extracts the installer and
// starts it. On Windows the app quits so the installer can replace the binary;
// elsewhere it opens the release page instead.
func (a *App) InstallAppUpdate() (AppUpdateDTO, error) {
	dto := a.CheckAppUpdate()
	if !dto.Available {
		return dto, nil
	}
	if runtime.GOOS != "windows" {
		if err := application.Get().Browser.OpenURL(releasesLatest); err != nil {
			return dto, err
		}
		dto.Detail = "Se abrió la página de descargas"
		return dto, nil
	}

	data, err := downloadRelease(updateClient, releasesDownload+dto.Latest+"/"+windowsAsset)
	if err != nil {
		return dto, err
	}
	dir, err := os.MkdirTemp("", "stefani-update-")
	if err != nil {
		return dto, err
	}
	installer := filepath.Join(dir, installerName)
	if err := extractAsset(data, installerName, installer); err != nil {
		return dto, err
	}
	if err := launchInstaller(installer); err != nil {
		return dto, fmt.Errorf("abrir el instalador: %w", err)
	}
	dto.Ready = true
	dto.Detail = "El instalador se abrió; la app se cerrará"
	go func() {
		time.Sleep(1200 * time.Millisecond)
		application.Get().Quit()
	}()
	return dto, nil
}

// latestTag reads the tag from the /releases/latest redirect instead of the
// API, which rate-limits shared IPs.
func latestTag(client *http.Client) (string, error) {
	req, err := http.NewRequest(http.MethodGet, releasesLatestURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "stefani")
	noRedirect := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	res, err := noRedirect.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	location := res.Header.Get("Location")
	if location == "" {
		return "", errors.New("sin redirección de la última versión")
	}
	location = strings.TrimSuffix(location, "/")
	tag := location[strings.LastIndex(location, "/")+1:]
	if tag == "" {
		return "", errors.New("tag vacío")
	}
	return tag, nil
}

// isNewer compares vX.Y.Z tags; an unstamped build ("dev") never prompts.
func isNewer(latest, current string) bool {
	next, ok := versionParts(latest)
	if !ok {
		return false
	}
	have, ok := versionParts(current)
	if !ok {
		return false
	}
	for i := 0; i < len(next) || i < len(have); i++ {
		l, c := 0, 0
		if i < len(next) {
			l = next[i]
		}
		if i < len(have) {
			c = have[i]
		}
		if l != c {
			return l > c
		}
	}
	return false
}

func versionParts(raw string) ([]int, bool) {
	raw = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(raw), "v"))
	if raw == "" || raw == "dev" {
		return nil, false
	}
	fields := strings.Split(raw, ".")
	parts := make([]int, 0, len(fields))
	for _, field := range fields {
		if idx := strings.IndexAny(field, "-+"); idx >= 0 {
			field = field[:idx]
		}
		n, err := strconv.Atoi(field)
		if err != nil {
			return nil, false
		}
		parts = append(parts, n)
	}
	return parts, len(parts) > 0
}

func downloadRelease(client *http.Client, url string) ([]byte, error) {
	res, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("descargar la actualización: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("descargar la actualización: HTTP %d", res.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(res.Body, 300<<20))
	if err != nil {
		return nil, err
	}
	if len(data) < 1<<20 {
		return nil, errors.New("la descarga de la actualización quedó incompleta")
	}
	return data, nil
}

func extractAsset(data []byte, name, target string) error {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return fmt.Errorf("leer el paquete: %w", err)
	}
	for _, file := range reader.File {
		if file.Name != name {
			continue
		}
		src, err := file.Open()
		if err != nil {
			return err
		}
		defer src.Close()
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, src); err != nil {
			out.Close()
			return err
		}
		return out.Close()
	}
	return fmt.Errorf("%s no está en el paquete", name)
}
