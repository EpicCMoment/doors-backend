package scripts

import (
	"strings"
	"testing"
)

func TestNamesContainsExpected(t *testing.T) {
	want := []string{"ping", "get_module", "list_items", "get_requirements", "get_baselines", "traverse_hierarchy"}
	names := Names()
	for _, w := range want {
		found := false
		for _, n := range names {
			if n == w {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing script %q in %v", w, names)
		}
	}
}

func TestLoadPrependsPreamble(t *testing.T) {
	s, err := Load("ping")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(s, "jsonEscape") {
		t.Error("preamble not prepended")
	}
	if !strings.Contains(s, "pong") {
		t.Error("body missing")
	}
}

func TestLoadUnknown(t *testing.T) {
	if _, err := Load("nope"); err == nil {
		t.Error("expected error")
	}
}
