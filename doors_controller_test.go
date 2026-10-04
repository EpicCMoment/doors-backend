package backend

import "testing"

func TestDefaultBeforeInit(t *testing.T) {
	if _, err := Default(); err == nil {
		t.Error("expected error before Init")
	}
}

func TestInitFailureKeepsDefaultNil(t *testing.T) {
	cfg := DefaultConfig()
	cfg.DoorsPath = "/nonexistent/doors"
	cfg.StartupTimeoutSec = 1
	if _, err := Init(cfg); err == nil {
		t.Fatal("expected startup error")
	}
	if _, err := Default(); err == nil {
		t.Error("expected Default to still fail after failed Init")
	}
}
