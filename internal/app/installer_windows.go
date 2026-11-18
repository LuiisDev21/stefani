//go:build windows

package app

import (
	"fmt"
	"path/filepath"
	"syscall"
	"unsafe"
)

var procShellExecuteW = syscall.NewLazyDLL("shell32.dll").NewProc("ShellExecuteW")

// launchInstaller starts the NSIS installer with the runas verb. The installer
// requests administrator rights, so CreateProcess (what exec.Command uses)
// fails with ERROR_ELEVATION_REQUIRED; only ShellExecute raises the UAC prompt.
func launchInstaller(path string) error {
	verb, err := syscall.UTF16PtrFromString("runas")
	if err != nil {
		return err
	}
	file, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	dir, err := syscall.UTF16PtrFromString(filepath.Dir(path))
	if err != nil {
		return err
	}
	const swShowNormal = 1
	ret, _, _ := procShellExecuteW.Call(
		0,
		uintptr(unsafe.Pointer(verb)),
		uintptr(unsafe.Pointer(file)),
		0,
		uintptr(unsafe.Pointer(dir)),
		swShowNormal,
	)
	if ret <= 32 {
		return fmt.Errorf("no se pudo iniciar el instalador (código %d)", ret)
	}
	return nil
}
