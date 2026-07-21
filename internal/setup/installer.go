package setup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// InstallMCPConfig configures phanes-dna MCP stdio entry into target agent config files.
func InstallMCPConfig(target string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get user home dir: %w", err)
	}

	execPath, err := os.Executable()
	if err != nil || execPath == "" {
		execPath = "phanes-dna"
	}

	configs := map[string]string{
		"claude":      filepath.Join(home, ".config", "Claude", "claude_desktop_config.json"),
		"cursor":      filepath.Join(home, ".cursor", "mcp.json"),
		"antigravity": filepath.Join(home, ".gemini", "antigravity-cli", "mcp.json"),
		"opencode":    filepath.Join(home, ".config", "opencode", "opencode.json"),
	}

	if target == "" {
		target = "all"
	}

	installedCount := 0
	for agent, path := range configs {
		if target != "all" && target != agent {
			continue
		}

		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			continue
		}

		var cfg map[string]any
		data, err := os.ReadFile(path)
		if err == nil {
			_ = json.Unmarshal(data, &cfg)
		}
		if cfg == nil {
			cfg = make(map[string]any)
		}

		mcpServers, ok := cfg["mcpServers"].(map[string]any)
		if !ok {
			mcpServers = make(map[string]any)
		}

		mcpServers["phanes-dna"] = map[string]any{
			"command": execPath,
			"args":    []string{"serve"},
		}

		cfg["mcpServers"] = mcpServers

		out, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			continue
		}

		if err := os.WriteFile(path, out, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "  ⚠️ Failed to write config for '%s': %v\n", agent, err)
			continue
		}

		fmt.Printf("  ✅ Configured MCP server for '%s' -> %s\n", agent, path)
		installedCount++
	}

	if installedCount == 0 {
		return fmt.Errorf("no agent config file updated for target %q", target)
	}

	return nil
}
