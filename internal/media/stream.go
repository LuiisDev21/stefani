package media

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// ErrNoAudio means the track has nothing to stream: no local file and no
// resolvable online source.
var ErrNoAudio = errors.New("no audio for this track")

// Source is where a track's bytes live: either a file on disk or a remote URL
// that this server proxies.
type Source struct {
	Path string
	URL  string
}

// ResolveFunc maps a track id from the request to a source.
type ResolveFunc func(r *http.Request, trackID string) (Source, error)

// StreamHandler serves audio to the webview with range support, which is what
// lets the player seek instead of downloading a whole file.
//
// Security: the URL only carries an internal track id, never a path. The
// resolver looks the id up in the database, so a crafted request cannot read
// arbitrary files, and remote URLs are limited to Google's media hosts.
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

		src, err := resolve(r, trackID)
		if err != nil {
			status := http.StatusNotFound
			if !errors.Is(err, ErrNoAudio) && !errors.Is(err, os.ErrNotExist) {
				status = http.StatusBadRequest
			}
			http.Error(w, "audio not available", status)
			return
		}
		serveSource(w, r, src)
	})
}

func serveSource(w http.ResponseWriter, r *http.Request, src Source) {
	switch {
	case src.Path != "":
		serveFile(w, r, src.Path)
	case src.URL != "":
		proxyRemote(w, r, src.URL)
	default:
		http.Error(w, "audio not available", http.StatusNotFound)
	}
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

// relayClient streams remote audio without a whole-body timeout; only the
// response headers are bounded, so a long song never gets cut.
var relayClient = &http.Client{
	Transport: &http.Transport{
		Proxy:                 nil,
		ResponseHeaderTimeout: 20 * time.Second,
		DisableCompression:    true,
		ForceAttemptHTTP2:     true,
		MaxIdleConnsPerHost:   4,
		IdleConnTimeout:       60 * time.Second,
	},
}

// proxyRemote relays a range request to the upstream media host. Proxying keeps
// the expiring, IP-bound media URL out of the webview and avoids CORS.
func proxyRemote(w http.ResponseWriter, r *http.Request, rawURL string) {
	target, err := url.Parse(rawURL)
	if err != nil || target.Scheme != "https" || !allowedRemoteHost(target.Hostname()) {
		http.Error(w, "audio not available", http.StatusNotFound)
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), r.Method, target.String(), nil)
	if err != nil {
		http.Error(w, "audio not available", http.StatusBadGateway)
		return
	}
	if rg := r.Header.Get("Range"); rg != "" {
		req.Header.Set("Range", rg)
	}
	if ir := r.Header.Get("If-Range"); ir != "" {
		req.Header.Set("If-Range", ir)
	}

	resp, err := relayClient.Do(req)
	if err != nil {
		http.Error(w, "audio not available", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for _, header := range []string{"Content-Type", "Content-Length", "Content-Range", "Accept-Ranges", "Last-Modified", "ETag"} {
		if value := resp.Header.Get(header); value != "" {
			w.Header().Set(header, value)
		}
	}
	w.Header().Set("Cache-Control", "private, no-store")
	w.WriteHeader(resp.StatusCode)
	if r.Method == http.MethodHead || resp.StatusCode == http.StatusRequestedRangeNotSatisfiable {
		return
	}
	_, _ = io.Copy(w, resp.Body)
}

// allowedRemoteHost limits relayed URLs to Google's media hosts. It is a
// variable so the proxy test can point it at a local TLS server.
var allowedRemoteHost = func(host string) bool {
	host = strings.ToLower(strings.TrimSuffix(host, "."))
	return host == "googlevideo.com" || strings.HasSuffix(host, ".googlevideo.com")
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
	case ".webm":
		return "audio/webm"
	}
	return mime.TypeByExtension(filepath.Ext(filePath))
}
