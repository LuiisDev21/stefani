//go:build windows

package app

import (
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

// appUserModelID is what Windows shows as the app name in the media flyout and
// the notification centre. Without an explicit id the shell falls back to
// "Unknown application".
const appUserModelID = "app.stefani.desktop"

// registerShellIdentity gives the process an identity the shell can resolve.
func registerShellIdentity() {
	proc := syscall.NewLazyDLL("shell32.dll").NewProc("SetCurrentProcessExplicitAppUserModelID")
	if id, err := syscall.UTF16PtrFromString(appUserModelID); err == nil {
		_, _, _ = proc.Call(uintptr(unsafe.Pointer(id)))
	}

	exe, err := os.Executable()
	if err != nil {
		return
	}
	key, _, err := registry.CreateKey(
		registry.CURRENT_USER,
		`Software\Classes\AppUserModelId\`+appUserModelID,
		registry.WRITE,
	)
	if err != nil {
		return
	}
	defer key.Close()
	_ = key.SetStringValue("DisplayName", "STEFANI")
	_ = key.SetStringValue("IconUri", exe)
}
