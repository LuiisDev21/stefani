//go:build !windows

package app

import "os/exec"

// launchInstaller starts the downloaded installer on systems without UAC.
func launchInstaller(path string) error {
	return exec.Command(path).Start()
}
