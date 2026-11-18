package app

import (
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// healthState tracks the frontend heartbeat so the host can recover a window
// whose renderer died or whose React tree unmounted and left it black.
type healthState struct {
	mu     sync.Mutex
	last   time.Time
	report string
}

// ClientHeartbeat is called by the UI every few seconds with a small JSON
// snapshot of the document (root children, visible text, captured errors).
func (a *App) ClientHeartbeat(report string) {
	a.health.mu.Lock()
	a.health.last = time.Now()
	a.health.report = report
	a.health.mu.Unlock()
}

// WatchRenderer reloads the window when the UI stops reporting or reports an
// empty tree. It never acts while the window is hidden or minimised, where the
// browser throttles timers and a pause is expected.
func (a *App) WatchRenderer(window *application.WebviewWindow, logger *slog.Logger) {
	a.health.mu.Lock()
	if a.health.last.IsZero() {
		a.health.last = time.Now()
	}
	a.health.mu.Unlock()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if !window.IsVisible() || window.IsMinimised() {
			continue
		}
		a.health.mu.Lock()
		last, report := a.health.last, a.health.report
		a.health.mu.Unlock()

		empty := strings.Contains(report, `"root":0`) && strings.Contains(report, `"text":0`)
		if !empty && time.Since(last) < 8*time.Second {
			continue
		}
		logger.Warn("la interfaz dejó de responder; recargando", "detalle", report)
		window.Reload()
		a.health.mu.Lock()
		a.health.last = time.Now()
		a.health.report = ""
		a.health.mu.Unlock()
	}
}
