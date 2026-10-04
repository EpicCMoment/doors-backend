package dxl

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// Controller owns the TCP conversation with the DXL server running inside
// DOORS. The server handles one request per connection, so each Exec opens
// a fresh connection; a mutex serializes concurrent callers.
type Controller struct {
	host string
	port int

	mu      sync.Mutex
	timeout time.Duration
}

// NewController creates a controller targeting the given host/port.
func NewController(host string, port int) *Controller {
	return &Controller{host: host, port: port, timeout: 60 * time.Second}
}

// SetTimeout overrides the per-request read timeout.
func (c *Controller) SetTimeout(d time.Duration) { c.timeout = d }

// Exec sends a rendered DXL script to the server and returns the raw JSON
// reply (the script's final string expression).
func (c *Controller) Exec(script string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", c.host, c.port), 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("dial DXL server: %w", err)
	}
	defer conn.Close()
	_ = conn.SetReadDeadline(time.Now().Add(c.timeout))
	_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	if _, err := conn.Write([]byte(script)); err != nil {
		return nil, fmt.Errorf("send script: %w", err)
	}
	reply, err := io.ReadAll(conn)
	if err != nil {
		return nil, fmt.Errorf("read reply: %w", err)
	}
	reply = bytes.TrimRight(reply, "\x00 \t\r\n")
	if len(reply) == 0 {
		return nil, fmt.Errorf("empty reply from DXL server")
	}
	return reply, nil
}

// Shutdown asks the DXL server to stop and waits for its acknowledgement.
func (c *Controller) Shutdown() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", c.host, c.port), 10*time.Second)
	if err != nil {
		return fmt.Errorf("dial DXL server: %w", err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(10 * time.Second))

	if _, err := conn.Write([]byte("shutdown_")); err != nil {
		return fmt.Errorf("send shutdown: %w", err)
	}
	reply, err := io.ReadAll(conn)
	if err != nil {
		return fmt.Errorf("read shutdown reply: %w", err)
	}
	if string(reply) != "done_" {
		return fmt.Errorf("unexpected shutdown reply %q", string(reply))
	}
	return nil
}
