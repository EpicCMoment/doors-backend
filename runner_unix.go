//go:build !windows

package backend

import "os/exec"

// applyPlatformAttrs is a no-op on non-Windows platforms.
func applyPlatformAttrs(cmd *exec.Cmd) {}
