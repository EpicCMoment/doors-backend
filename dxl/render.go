package dxl

import (
	"bytes"
	_ "embed"
	"fmt"
	"strconv"
	"strings"
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

// DxlLiteral converts a Go value into a DXL source literal for safe
// injection into templated scripts. The concrete Go type determines the
// literal form: strings are quoted/escaped, integers are plain, booleans
// become DXL true/false, and raw strings are spliced in verbatim.
type DxlLiteral struct {
	literalKind DxlLiteralKind
	text        string
	integer     int64
	boolean     bool
}

type DxlLiteralKind int

const (
	dxlString DxlLiteralKind = iota
	dxlInteger
	dxlBoolean
	dxlRaw
)

func String(s string) DxlLiteral { return DxlLiteral{literalKind: dxlString, text: s} }
func Int(n int64) DxlLiteral     { return DxlLiteral{literalKind: dxlInteger, integer: n} }
func Bool(b bool) DxlLiteral     { return DxlLiteral{literalKind: dxlBoolean, boolean: b} }

// Raw splices pre-validated DXL source directly into the script.
func Raw(src string) DxlLiteral { return DxlLiteral{literalKind: dxlRaw, text: src} }

// Literal renders the DXL source form of the value.
func (l DxlLiteral) Literal() string {
	switch l.literalKind {
	case dxlRaw:
		return l.text
	case dxlString:
		return strconv.Quote(l.text) // quoted, with backslash escapes
	case dxlInteger:
		return strconv.FormatInt(l.integer, 10)
	case dxlBoolean:
		if l.boolean {
			return "true"
		}
		return "false"
	}
	return ""
}

// Render executes a script template, injecting each parameter as its DXL
// literal (e.g. {{.modulePath}} becomes "\"/MyProject/Mod\"").
func Render(tplText string, params map[string]DxlLiteral) (string, error) {
	tpl, err := template.New("script").Parse(tplText)
	if err != nil {
		return "", fmt.Errorf("parse script template: %w", err)
	}
	data := map[string]string{}
	for k, v := range params {
		data[k] = v.Literal()
	}
	var buf strings.Builder
	if err := tpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("render script: %w", err)
	}
	return buf.String(), nil
}
