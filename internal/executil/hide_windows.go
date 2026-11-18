//go:build windows

package executil

import (
	"os/exec"
	"syscall"
)

// createNoWindow stops a console child from allocating a console at all.
const createNoWindow = 0x08000000

// Hide keeps console children (yt-dlp, the login process) from opening a black
// console window over the app.
func Hide(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.HideWindow = true
	cmd.SysProcAttr.CreationFlags |= createNoWindow
}
