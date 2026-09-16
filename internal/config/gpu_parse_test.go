package config

import (
	"os"
	"testing"
)

func TestGPUParseRoundTrip(t *testing.T) {
	f := t.TempDir() + "/fence.json"
	if err := os.WriteFile(f, []byte(`{"macos": {"gpu": true}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	c, err := Load(f)
	if err != nil {
		t.Fatal(err)
	}
	if c.MacOS.Gpu == nil || !*c.MacOS.Gpu {
		t.Fatalf("gpu = %v, want true", c.MacOS.Gpu)
	}
	merged := Merge(Default(), c)
	if merged.MacOS.Gpu == nil || !*merged.MacOS.Gpu {
		t.Fatalf("merged gpu = %v, want true", merged.MacOS.Gpu)
	}
}
