//go:build windows

package backend

import (
	"os/exec"
	"syscall"
)

// applyPlatformAttrs hides the DOORS console window on Windows.
func applyPlatformAttrs(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}
