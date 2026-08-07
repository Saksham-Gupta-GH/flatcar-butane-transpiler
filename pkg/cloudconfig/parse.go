// Package cloudconfig provides utilities for parsing cloud-init cloud-config
// YAML documents.
package cloudconfig

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Parse reads and parses a cloud-config YAML file from the given path.
func Parse(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading cloud-config file: %w", err)
	}
	return ParseBytes(data)
}

// ParseBytes parses cloud-config YAML from a byte slice.
// It strips the optional "#cloud-config" header line before parsing.
func ParseBytes(data []byte) (*Config, error) {
	content := string(data)

	// Strip the #cloud-config shebang-like header if present
	if strings.HasPrefix(strings.TrimSpace(content), "#cloud-config") {
		lines := strings.SplitN(content, "\n", 2)
		if len(lines) == 2 {
			content = lines[1]
		} else {
			content = ""
		}
	}

	var cfg Config
	if err := yaml.Unmarshal([]byte(content), &cfg); err != nil {
		return nil, fmt.Errorf("parsing cloud-config YAML: %w", err)
	}
	return &cfg, nil
}
