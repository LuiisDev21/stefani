//go:build !windows

package app

// registerShellIdentity is a no-op outside Windows, where the desktop
// environment names the app from the bundle instead.
func registerShellIdentity() {}
