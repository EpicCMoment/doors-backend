// Command launch is a small interactive driver that exercises doors-backend
// as a library: it starts the DOORS DXL server (via the embedded script) and
// lets you issue backend calls from a simple read-eval-print loop.
//
// Usage:
//
//	doors-backend launch <config.json>
//
// Config: see doors-backend README. The DoorsPath in the config must point at
// a real DOORS 9.7 executable.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	backend "github.com/EpicCMoment/doors-backend"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: launch <config.json>")
		os.Exit(2)
	}
	cfg, err := backend.LoadConfig(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "config:", err)
		os.Exit(1)
	}

	d, err := backend.Init(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "backend init:", err)
		os.Exit(1)
	}
	defer d.Shutdown()

	fmt.Printf("DOORS running on %s:%d. Type 'help', 'quit' to exit.\n", cfg.Host, cfg.Port)
	sc := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !sc.Scan() {
			break
		}
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		if cmd, err := run(line, d); err != nil {
			fmt.Println("error:", err)
		} else {
			fmt.Println(cmd)
		}
	}
}

func run(line string, d *backend.DoorsController) (string, error) {
	parts := strings.Fields(line)
	switch parts[0] {
	case "help":
		return "commands: ping | module <path> [baseline] | mods <path> | reqs <modulePath> [baseline] | baselines <modulePath> | hierarchy <modulePath> [baseline] | quit", nil
	case "quit", "exit":
		os.Exit(0)
		return "", nil
	case "ping":
		raw, err := d.Ping()
		return string(raw), err
	case "module":
		if len(parts) < 2 {
			return "", fmt.Errorf("usage: module <path> [baseline]")
		}
		bl := ""
		if len(parts) > 2 {
			bl = parts[2]
		}
		m, err := d.Modules().GetModule(parts[1], bl)
		return pretty(m), err
	case "mods":
		if len(parts) < 2 {
			return "", fmt.Errorf("usage: mods <path>")
		}
		mods, err := d.Modules().ListModules(parts[1])
		return pretty(mods), err
	case "reqs":
		if len(parts) < 2 {
			return "", fmt.Errorf("usage: reqs <modulePath> [baseline]")
		}
		bl := ""
		if len(parts) > 2 {
			bl = parts[2]
		}
		reqs, err := d.Modules().GetRequirements(parts[1], bl)
		return pretty(reqs), err
	case "baselines":
		if len(parts) < 2 {
			return "", fmt.Errorf("usage: baselines <modulePath>")
		}
		bs, err := d.Modules().GetBaselines(parts[1])
		return pretty(bs), err
	case "hierarchy":
		if len(parts) < 2 {
			return "", fmt.Errorf("usage: hierarchy <modulePath> [baseline]")
		}
		bl := ""
		if len(parts) > 2 {
			bl = parts[2]
		}
		h, err := d.Modules().TraverseHierarchy(parts[1], bl)
		return pretty(h), err
	default:
		return "", fmt.Errorf("unknown command %q (try help)", parts[0])
	}
}

func pretty(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}
