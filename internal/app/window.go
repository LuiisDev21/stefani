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

func (a *App) WindowClose() {
	if w := application.Get().Window.Current(); w != nil {
		w.Close()
	}
}
