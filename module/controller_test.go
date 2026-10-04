package module

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/EpicCMoment/doors-backend/dxl"
)

type fakeExec struct {
	lastName string
	lastArgs map[string]dxl.DxlLiteral
	reply    []byte
	err      error
}

func (f *fakeExec) ExecTemplate(name string, params map[string]dxl.DxlLiteral) ([]byte, error) {
	f.lastName = name
	f.lastArgs = params
	return f.reply, f.err
}

func okReply(t *testing.T, data json.RawMessage) []byte {
	t.Helper()
	env := dxl.Envelope{Status: "ok", Data: data}
	raw, err := json.Marshal(env)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestGetModuleFetchesAndCaches(t *testing.T) {
	data, _ := json.Marshal(Module{ID: "1", Name: "M", Path: "/P/M"})
	f := &fakeExec{reply: okReply(t, data)}
	c := NewModuleController(f)

	m, err := c.GetModule("/P/M", "")
	if err != nil {
		t.Fatal(err)
	}
	if m.Name != "M" || f.lastName != "get_module" {
		t.Fatalf("unexpected: %+v lastName=%s", m, f.lastName)
	}
	if lit := f.lastArgs["path"].Literal(); lit != `"/P/M"` {
		t.Errorf("path literal = %s", lit)
	}

	// second call should hit the cache, not ExecTemplate
	f.reply = nil
	m2, err := c.GetModule("/P/M", "")
	if err != nil || m2.Name != "M" {
		t.Fatalf("cache miss: %v %+v", err, m2)
	}
}

func TestControllerErrorEnvelope(t *testing.T) {
	env := dxl.Envelope{Status: "error", Message: "boom"}
	raw, _ := json.Marshal(env)
	f := &fakeExec{reply: raw}
	c := NewModuleController(f)
	if _, err := c.GetModule("/P/M", ""); err == nil || !strings.Contains(err.Error(), "boom") {
		t.Errorf("expected boom error, got %v", err)
	}
}

func TestControllerTransportError(t *testing.T) {
	f := &fakeExec{err: errors.New("conn refused")}
	c := NewModuleController(f)
	if _, err := c.ListModules("/P"); err == nil {
		t.Error("expected error")
	}
}

func TestGetBaselinesCaches(t *testing.T) {
	bs := []Baseline{{ModuleID: "1", Name: "b1"}}
	data, _ := json.Marshal(bs)
	f := &fakeExec{reply: okReply(t, data)}
	c := NewModuleController(f)
	got, err := c.GetBaselines("1")
	if err != nil || len(got) != 1 {
		t.Fatalf("%v %+v", err, got)
	}
	f.reply = nil
	got, _ = c.GetBaselines("1")
	if len(got) != 1 {
		t.Fatal("cache miss for baselines")
	}
}
