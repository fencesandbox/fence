package fence

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMergeConfigs(t *testing.T) {
	base := &Config{
		Network: NetworkConfig{
			AllowedDomains: []string{"example.com"},
		},
	}
	override := &Config{
		Extends: "base-template",
		Network: NetworkConfig{
			AllowedDomains: []string{"api.example.com"},
		},
	}

	result := MergeConfigs(base, override)
	if result == nil {
		t.Fatal("expected non-nil merged config")
	}
	if result.Extends != "" {
		t.Fatalf("expected Extends to be cleared, got %q", result.Extends)
	}
	if len(result.Network.AllowedDomains) != 2 {
		t.Fatalf("expected 2 allowed domains, got %d: %v", len(result.Network.AllowedDomains), result.Network.AllowedDomains)
	}
}

func TestLoadConfigResolved(t *testing.T) {
	dir := t.TempDir()
	basePath := filepath.Join(dir, "base.json")
	overridePath := filepath.Join(dir, "override.json")

	if err := os.WriteFile(basePath, []byte(`{
  "network": {
    "allowedDomains": ["example.com"]
  }
}`), 0o600); err != nil {
		t.Fatalf("write base config: %v", err)
	}

	if err := os.WriteFile(overridePath, []byte(`{
  "extends": "./base.json",
  "filesystem": {
    "allowWrite": [".tmp"]
  }
}`), 0o600); err != nil {
		t.Fatalf("write override config: %v", err)
	}

	resolved, err := LoadConfigResolved(overridePath)
	if err != nil {
		t.Fatalf("load resolved config: %v", err)
	}

	if resolved == nil {
		t.Fatal("expected resolved config")
	}
	if resolved.Extends != "" {
		t.Fatalf("expected Extends to be cleared, got %q", resolved.Extends)
	}
	if len(resolved.Network.AllowedDomains) != 1 || resolved.Network.AllowedDomains[0] != "example.com" {
		t.Fatalf("expected inherited allowed domain, got %v", resolved.Network.AllowedDomains)
	}
	if len(resolved.Filesystem.AllowWrite) != 1 || resolved.Filesystem.AllowWrite[0] != ".tmp" {
		t.Fatalf("expected allowWrite to be preserved, got %v", resolved.Filesystem.AllowWrite)
	}
}
