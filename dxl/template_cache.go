package dxl

import (
	"fmt"
	"strings"
	"sync"
	"text/template"
)

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
