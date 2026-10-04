package dxl

import (
	"reflect"
	"strings"
	"testing"
)

func TestDxlLiteralForms(t *testing.T) {
	cases := []struct {
		name string
		lit  DxlLiteral
		want string
	}{
		{"string", String(`/MyProject/Mod"s`), `"/MyProject/Mod\"s"`},
		{"integer", Int(42), "42"},
		{"negative integer", Int(-7), "-7"},
		{"bool true", Bool(true), "true"},
		{"bool false", Bool(false), "false"},
		{"raw", Raw(`"x"+y`), `"x"+y`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.lit.Literal(); got != tc.want {
				t.Errorf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestRenderInjectsTypedLiterals(t *testing.T) {
	tpl := "Module m = read({{.path}}, false)\nint n = {{.count}}\nbool b = {{.flag}}"
	got, err := Render(tpl, map[string]DxlLiteral{
		"path":  String("/P/M"),
		"count": Int(3),
		"flag":  Bool(true),
	})
	if err != nil {
		t.Fatal(err)
	}
	want := "Module m = read(\"/P/M\", false)\nint n = 3\nbool b = true"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestTemplateCacheRender(t *testing.T) {
	c := NewTemplateCache()
	if err := c.Register("t", "hello {{.name}}"); err != nil {
		t.Fatal(err)
	}
	got, err := c.Render("t", map[string]DxlLiteral{"name": String("doors")})
	if err != nil {
		t.Fatal(err)
	}
	if got != `hello "doors"` {
		t.Errorf("got %q", got)
	}
	if _, err := c.Render("missing", nil); err == nil || !strings.Contains(err.Error(), "not registered") {
		t.Errorf("expected not registered error, got %v", err)
	}
}

func TestRenderServerScript(t *testing.T) {
	src, err := RenderServerScript(61610)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(src, "server 61610") {
		t.Errorf("port not injected: %s", src[:120])
	}
	if strings.Contains(src, "{{") {
		t.Errorf("unrendered placeholder remains")
	}
}

func TestExecuteReusedTemplatesProduceSameResult(t *testing.T) {
	c := NewTemplateCache()
	c.MustRegister("t", "{{.x}}")
	a, _ := c.Render("t", map[string]DxlLiteral{"x": Int(1)})
	b, _ := c.Render("t", map[string]DxlLiteral{"x": Int(1)})
	if !reflect.DeepEqual(a, b) {
		t.Errorf("%q != %q", a, b)
	}
}
