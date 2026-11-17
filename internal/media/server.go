package media

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

// Server streams local audio over a loopback HTTP listener.
//
// Why a socket instead of the asset server: the player needs byte ranges (to
// seek) and the same URL must work in `wails3 dev`, in the packaged app and in
// the browser E2E build. A loopback listener with a per-session token gives all
// three without exposing anything beyond 127.0.0.1.
type Server struct {
	listener net.Listener
	token    string
	baseURL  string
	resolve  ResolveFunc
	http     *http.Server
}

// NewServer starts the loopback listener on a random port.
func NewServer(resolve ResolveFunc) (*Server, error) {
	if resolve == nil {
		return nil, fmt.Errorf("media server needs a resolver")
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("start media server: %w", err)
	}
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		listener.Close()
		return nil, fmt.Errorf("generate media token: %w", err)
	}
	s := &Server{
		listener: listener,
		token:    hex.EncodeToString(raw[:]),
		baseURL:  "http://" + listener.Addr().String(),
		resolve:  resolve,
	}
	mux := http.NewServeMux()
	mux.Handle("/media/", s.authorized(http.HandlerFunc(s.serve)))
	s.http = &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		_ = s.http.Serve(listener)
	}()
	return s, nil
}

// BaseURL is the loopback origin, used by diagnostics and tests.
func (s *Server) BaseURL() string { return s.baseURL }

// URL returns a playable URL for a track id.
func (s *Server) URL(trackID string) string {
	return s.baseURL + "/media/" + trackID + "?token=" + s.token
}

// Close stops the listener.
func (s *Server) Close() error {
	if s.http != nil {
		s.http.Close()
	}
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *Server) authorized(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("token") != s.token {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		// The webview origin differs between dev and production; media loading
		// does not need credentials, so a permissive origin is safe here.
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Range")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Range, Accept-Ranges, Content-Length")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) serve(w http.ResponseWriter, r *http.Request) {
	trackID := strings.TrimPrefix(r.URL.Path, "/media/")
	if trackID == "" || strings.Contains(trackID, "..") || strings.Contains(trackID, "/") {
		http.Error(w, "invalid media id", http.StatusBadRequest)
		return
	}
	src, err := s.resolve(r, trackID)
	if err != nil {
		http.Error(w, "audio not available", http.StatusNotFound)
		return
	}
	serveSource(w, r, src)
}
