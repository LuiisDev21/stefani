package media

import (
	"errors"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// ErrNoAudio means the track has no local audio to stream.
var ErrNoAudio = errors.New("no audio for this track")

// ResolveFunc maps a track id from the request to a file path on disk.
type ResolveFunc func(r *http.Request, trackID string) (string, error)

// StreamHandler serves local audio to the webview with range support, which is
// what lets the player seek instead of downloading a whole file.
//
// Security: the URL only carries an internal track id, never a path. The
// resolver is expected to look the id up in the database, so a crafted request
// cannot read arbitrary files.
func StreamHandler(resolve ResolveFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/media/") {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		trackID := strings.TrimPrefix(r.URL.Path, "/media/")
		trackID = path.Clean("/" + trackID)
		trackID = strings.TrimPrefix(trackID, "/")
		if trackID == "" || strings.Contains(trackID, "..") {
			http.Error(w, "invalid media id", http.StatusBadRequest)
			return
		}

		filePath, err := resolve(r, trackID)
		if err != nil {
			status := http.StatusNotFound
			if !errors.Is(err, ErrNoAudio) && !errors.Is(err, os.ErrNotExist) {
				status = http.StatusBadRequest
			}
			http.Error(w, "audio not available", status)
			return
		}
		serveFile(w, r, filePath)
	})
}

func serveFile(w http.ResponseWriter, r *http.Request, filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "audio not available", http.StatusNotFound)
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || info.IsDir() {
		http.Error(w, "audio not available", http.StatusNotFound)
		return
	}
	if ct := contentType(filePath); ct != "" {
		w.Header().Set("Content-Type", ct)
	}
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Cache-Control", "private, max-age=3600")
	// ServeContent handles Range, HEAD and 206 responses.
	http.ServeContent(w, r, filepath.Base(filePath), info.ModTime(), file)
}

// contentType maps a file extension to a MIME type the webview can decode.
func contentType(filePath string) string {
	switch strings.ToLower(filepath.Ext(filePath)) {
	case ".wav":
		return "audio/wav"
	case ".mp3":
		return "audio/mpeg"
	case ".m4a", ".aac":
		return "audio/mp4"
	case ".flac":
		return "audio/flac"
	case ".ogg", ".opus":
		return "audio/ogg"
	}
	return mime.TypeByExtension(filepath.Ext(filePath))
}
