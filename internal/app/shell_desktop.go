//go:build !server

package app

// shellDesktop reports the native window: the desktop build owns its frame.
func shellDesktop() bool { return true }
