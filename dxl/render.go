package dxl

import (
	"bytes"
	_ "embed"
	"fmt"
	"strconv"
	"strings"
	"sync"
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
	return execute(tpl, params)
}

func execute(tpl *template.Template, params map[string]DxlLiteral) (string, error) {
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

// TemplateCache parses script templates once and reuses the parsed form,
// avoiding re-parsing on every request. Safe for concurrent use.
type TemplateCache struct {
	mu        sync.RWMutex
	templates map[string]*template.Template
}

// NewTemplateCache creates an empty cache.
func NewTemplateCache() *TemplateCache {
	return &TemplateCache{templates: map[string]*template.Template{}}
}

// Register parses and stores a template under the given name.
func (c *TemplateCache) Register(name, tplText string) error {
	tpl, err := template.New(name).Parse(tplText)
	if err != nil {
		return fmt.Errorf("parse template %q: %w", name, err)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.templates[name] = tpl
	return nil
}

// MustRegister panics on a parse error; useful for embedded templates at init.
func (c *TemplateCache) MustRegister(name, tplText string) {
	if err := c.Register(name, tplText); err != nil {
		panic(err)
	}
}

// Render executes the cached template with typed parameter injection.
func (c *TemplateCache) Render(name string, params map[string]DxlLiteral) (string, error) {
	c.mu.RLock()
	tpl, ok := c.templates[name]
	c.mu.RUnlock()
	if !ok {
		return "", fmt.Errorf("template %q not registered", name)
	}
	return execute(tpl, params)
}
