package dxl

import (
	"bytes"
	_ "embed"
	"fmt"
	"text/template"
)

//go:embed server/dxl_server.dxl
var serverScript string

// ServerScript returns the embedded DXL TCP server source, unrendered.
func ServerScript() string { return serverScript }

// RenderServerScript renders the server script with the given TCP port.
func RenderServerScript(port int) (string, error) {
	tpl, err := template.New("dxl_server").Parse(serverScript)
	if err != nil {
		return "", fmt.Errorf("parse embedded server script: %w", err)
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, map[string]int{"Port": port}); err != nil {
		return "", fmt.Errorf("render server script: %w", err)
	}
	return buf.String(), nil
}
