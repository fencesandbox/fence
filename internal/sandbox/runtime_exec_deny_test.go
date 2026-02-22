package sandbox

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/Use-Tusk/fence/internal/config"
)

func TestRuntimeExecutableToken(t *testing.T) {
	tests := []struct {
		rule string
		want string
		ok   bool
	}{
		{"python3", "python3", true},
		{" /usr/bin/python3 ", "/usr/bin/python3", true},
		{"git push", "", false},
		{"dd if=", "", false},
		{"python*", "", false},
		{"", "", false},
	}

	for _, tt := range tests {
		got, ok := runtimeExecutableToken(tt.rule)
		if ok != tt.ok || got != tt.want {
			t.Fatalf("runtimeExecutableToken(%q) = (%q,%v), want (%q,%v)", tt.rule, got, ok, tt.want, tt.ok)
		}
	}
}

func TestGetRuntimeDeniedExecutablePaths_SingleTokenOnly(t *testing.T) {
	cfg := &config.Config{
		Command: config.CommandConfig{
			Deny: []string{"python3", "git push", "dd if=", "bash -c"},
		},
	}

	got := GetRuntimeDeniedExecutablePaths(cfg)
	if len(resolveExecutablePaths("python3")) == 0 {
		t.Skip("python3 not available on this system")
	}
	if len(got) == 0 {
		t.Fatalf("expected at least one resolved path for single-token deny entry")
	}

	for _, p := range got {
		base := filepath.Base(p)
		if slices.Contains([]string{"git", "dd", "bash"}, base) {
			t.Fatalf("unexpected direct binary path in results: %s", p)
		}
	}
}
