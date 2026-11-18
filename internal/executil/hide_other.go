//go:build !windows

package executil

import "os/exec"

// Hide is a no-op outside Windows.
func Hide(*exec.Cmd) {}
