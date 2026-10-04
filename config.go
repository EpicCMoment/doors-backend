package backend

import (
	"encoding/json"
	"fmt"
	"os"
)

// DefaultDxlServerPort is the TCP port the embedded DXL server listens on.
const DefaultDxlServerPort = 61610

// Config holds every setting the backend needs. It is loaded from a JSON
// file by the MCP server and passed into the backend as a plain struct.
type Config struct {
	// DoorsPath is the absolute path to the DOORS executable
	// (doors.exe on Windows, bin/doors on Linux).
	DoorsPath string `json:"doorsPath"`

	// Host is the interface/host the DXL server binds to. Usually localhost.
	Host string `json:"host"`

	// Port is the TCP port the DXL server listens on.
	Port int `json:"port"`

	// Username / Password are passed to DOORS for database authentication.
	Username string `json:"username"`
	Password string `json:"password"`

	// StartupTimeoutSec bounds how long the runner waits for the DXL
	// server TCP port to accept connections.
	StartupTimeoutSec int `json:"startupTimeoutSec"`

	// KeepProcess keeps the DOORS process alive across DoorsController
	// Close calls when true. Mainly used for debugging.
	KeepProcess bool `json:"keepProcess"`
}

// DefaultConfig returns a Config populated with sane defaults.
func DefaultConfig() Config {
	return Config{
		Host:              "127.0.0.1",
		Port:              DefaultDxlServerPort,
		StartupTimeoutSec: 60,
		KeepProcess:       false,
	}
}

// LoadConfig reads a JSON config file and merges it over the defaults.
func LoadConfig(path string) (Config, error) {
	cfg := DefaultConfig()
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("read config: %w", err)
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config %s: %w", path, err)
	}
	if cfg.Port == 0 {
		cfg.Port = DefaultDxlServerPort
	}
	if cfg.StartupTimeoutSec <= 0 {
		cfg.StartupTimeoutSec = DefaultConfig().StartupTimeoutSec
	}
	if cfg.Host == "" {
		cfg.Host = DefaultConfig().Host
	}
	return cfg, nil
}
