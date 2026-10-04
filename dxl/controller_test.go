package dxl

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

// fakeDxlServer mimics the wire protocol of dxl_server.dxl: one request per
// connection, reply then close. handler maps request->reply.
func fakeDxlServer(t *testing.T, handler func(req string) string) (addr string, stop func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				br := bufio.NewReader(c)
				// read whatever the client sends before it waits for the reply
				_ = c.SetReadDeadline(time.Now().Add(2 * time.Second))
				buf := make([]byte, 64*1024)
				n, _ := br.Read(buf)
				req := string(buf[:n])
				reply := handler(req)
				_, _ = c.Write([]byte(reply))
			}(conn)
		}
	}()
	return ln.Addr().String(), func() { ln.Close() }
}

func TestControllerExecRoundTrip(t *testing.T) {
	addr, stop := fakeDxlServer(t, func(req string) string {
		if req == "ping" {
			return `{"status":"ok","data":{"pong":true}}`
		}
		return `{"status":"error","message":"unknown"}`
	})
	defer stop()

	host, portStr, _ := net.SplitHostPort(addr)
	var port int
	fmt.Sscanf(portStr, "%d", &port)

	c := NewController(host, port)
	reply, err := c.Exec("ping")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(reply), `"pong":true`) {
		t.Errorf("unexpected reply %s", reply)
	}
}

func TestControllerExecTemplate(t *testing.T) {
	addr, stop := fakeDxlServer(t, func(req string) string {
		if strings.Contains(req, `read("/P/M"`) {
			return `{"status":"ok"}`
		}
		return `{"status":"error","message":"bad render: " + req}`
	})
	defer stop()

	host, portStr, _ := net.SplitHostPort(addr)
	var port int
	fmt.Sscanf(portStr, "%d", &port)

	c := NewController(host, port)
	c.MustRegisterTemplate("get", `read({{.path}}, false)`)
	reply, err := c.ExecTemplate("get", map[string]DxlLiteral{"path": String("/P/M")})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(reply), `"ok"`) {
		t.Errorf("unexpected reply %s", reply)
	}
}

func TestControllerShutdown(t *testing.T) {
	addr, stop := fakeDxlServer(t, func(req string) string {
		if req == "shutdown_" {
			return "done_"
		}
		return ""
	})
	defer stop()

	host, portStr, _ := net.SplitHostPort(addr)
	var port int
	fmt.Sscanf(portStr, "%d", &port)

	c := NewController(host, port)
	if err := c.Shutdown(); err != nil {
		t.Fatal(err)
	}
}
