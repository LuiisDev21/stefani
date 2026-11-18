package app

import (
	"strings"
	"sync"
	"time"
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

// RendererStatus reports whether the UI stopped reporting or reported an empty
// tree. The host calls it from its watch loop; it stays a method so the loop
// does not reach into the app's internals.
func (a *App) RendererStatus() (bool, string) {
	a.health.mu.Lock()
	defer a.health.mu.Unlock()
	if a.health.last.IsZero() {
		a.health.last = time.Now()
		return false, ""
	}
	empty := strings.Contains(a.health.report, `"root":0`) && strings.Contains(a.health.report, `"text":0`)
	if !empty && time.Since(a.health.last) < 8*time.Second {
		return false, ""
	}
	return true, a.health.report
}

// ResetRendererHealth forgets the last report right before a reload.
func (a *App) ResetRendererHealth() {
	a.health.mu.Lock()
	a.health.last = time.Now()
	a.health.report = ""
	a.health.mu.Unlock()
}
