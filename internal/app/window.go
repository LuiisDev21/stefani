package app

import "github.com/wailsapp/wails/v3/pkg/application"

// WindowMinimise, WindowToggleMaximise and WindowClose back the controls the
// app draws in its own title bar; the window is frameless, so the system no
// longer provides them.
func (a *App) WindowMinimise() {
	if w := application.Get().Window.Current(); w != nil {
		w.Minimise()
	}
}

func (a *App) WindowToggleMaximise() {
	if w := application.Get().Window.Current(); w != nil {
		w.ToggleMaximise()
	}
}

// WindowSetFullscreen puts the window in real fullscreen, or restores it. The
// immersive view sets the state on mount and unmount instead of toggling, so a
// missed call cannot leave the window stuck without a title bar.
func (a *App) WindowSetFullscreen(fullscreen bool) {
	w := application.Get().Window.Current()
	if w == nil {
		return
	}
	if fullscreen {
		if !w.IsFullscreen() {
			w.Fullscreen()
		}
		return
	}
	if w.IsFullscreen() {
		w.UnFullscreen()
	}
}

func (a *App) WindowClose() {
	if w := application.Get().Window.Current(); w != nil {
		w.Close()
	}
}
