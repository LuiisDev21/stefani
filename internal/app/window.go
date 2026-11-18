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

// WindowToggleFullscreen takes the window to real fullscreen, not just the
// maximised size: the immersive view has no title bar to restore from.
func (a *App) WindowToggleFullscreen() {
	if w := application.Get().Window.Current(); w != nil {
		w.ToggleFullscreen()
	}
}

func (a *App) WindowClose() {
	if w := application.Get().Window.Current(); w != nil {
		w.Close()
	}
}
