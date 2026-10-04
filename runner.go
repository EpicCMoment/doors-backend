package backend

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/EpicCMoment/doors-backend/dxl"
)

// Runner spawns a background DOORS process executing the embedded DXL TCP
// server, waits for its port, and tears it down cleanly.
type Runner struct {
	cfg Config

	cmd        *exec.Cmd
	scriptPath string
}

// NewRunner creates a runner from the given config.
func NewRunner(cfg Config) *Runner { return &Runner{cfg: cfg} }

// Start launches DOORS in batch mode with the DXL server script and blocks
// until the server accepts TCP connections or the startup timeout elapses.
func (r *Runner) Start() error {
	script, err := dxl.RenderServerScript(r.cfg.Port)
	if err != nil {
		return err
	}
	dir, err := os.MkdirTemp("", "doors-dxl-server-*")
	if err != nil {
		return fmt.Errorf("temp dir: %w", err)
	}
	r.scriptPath = filepath.Join(dir, "dxl_server.dxl")
	if err := os.WriteFile(r.scriptPath, []byte(script), 0o600); err != nil {
		return fmt.Errorf("write server script: %w", err)
	}

	r.cmd = exec.Command(r.cfg.DoorsPath, "-batch", r.scriptPath)
	applyPlatformAttrs(r.cmd)
	if err := r.cmd.Start(); err != nil {
		return fmt.Errorf("start doors: %w", err)
	}
	return r.waitReady()
}

// waitReady polls the server port until it accepts a connection.
func (r *Runner) waitReady() error {
	timeout := time.Duration(r.cfg.StartupTimeoutSec) * time.Second
	deadline := time.Now().Add(timeout)
	addr := net.JoinHostPort(r.cfg.Host, fmt.Sprintf("%d", r.cfg.Port))
	for {
		conn, err := net.DialTimeout("tcp", addr, time.Second)
		if err == nil {
			conn.Close()
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("dxl server did not start within %s", timeout)
		}
		time.Sleep(250 * time.Millisecond)
	}
}

// Stop kills the DOORS process (unless KeepProcess is set) and removes the
// extracted server script. Callers can issue DxlController.Shutdown() first
// for a graceful stop.
func (r *Runner) Stop() error {
	if r.cfg.KeepProcess {
		return nil
	}
	if r.cmd != nil && r.cmd.Process != nil {
		_ = r.cmd.Process.Kill()
		_, _ = r.cmd.Process.Wait()
	}
	if r.scriptPath != "" {
		os.RemoveAll(filepath.Dir(r.scriptPath))
	}
	return nil
}
