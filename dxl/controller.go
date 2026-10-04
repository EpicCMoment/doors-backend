package dxl

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// DxlController owns the TCP conversation with the DXL server running inside
// DOORS. The server handles one request per connection, so each Exec opens
// a fresh connection; a mutex serializes concurrent callers.
type DxlController struct {
	host string
	port int

	mu      sync.Mutex
	timeout time.Duration

	templates *TemplateCache
}

// NewDxlController creates a controller targeting the given host/port.
func NewDxlController(host string, port int) *DxlController {
	return &DxlController{
		host:      host,
		port:      port,
		timeout:   60 * time.Second,
		templates: NewTemplateCache(),
	}
}

// RegisterTemplate parses and caches a script template under name.
func (c *DxlController) RegisterTemplate(name, tplText string) error {
	return c.templates.Register(name, tplText)
}

// MustRegisterTemplate panics on a parse error; handy at startup with embedded scripts.
func (c *DxlController) MustRegisterTemplate(name, tplText string) {
	c.templates.MustRegister(name, tplText)
}

// Template renders a cached script with typed parameter injection.
func (c *DxlController) Template(name string, params map[string]DxlLiteral) (string, error) {
	return c.templates.Render(name, params)
}

// SetTimeout overrides the per-request read timeout.
func (c *DxlController) SetTimeout(d time.Duration) { c.timeout = d }

// ExecTemplate renders a cached template and executes it on the DXL server.
func (c *DxlController) ExecTemplate(name string, params map[string]DxlLiteral) ([]byte, error) {
	script, err := c.Template(name, params)
	if err != nil {
		return nil, err
	}
	return c.Exec(script)
}

// Exec sends a rendered DXL script to the server and returns the raw JSON
// reply (the script's final string expression).
func (c *DxlController) Exec(script string) ([]byte, error) {
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
func (c *DxlController) Shutdown() error {
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
