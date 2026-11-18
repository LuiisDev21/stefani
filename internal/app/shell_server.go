//go:build server

package app

// shellDesktop is false in the HTTP build: there is no window to draw around.
func shellDesktop() bool { return false }
