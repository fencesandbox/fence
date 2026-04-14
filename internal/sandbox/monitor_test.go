package sandbox

import (
	"strings"
	"testing"
)

func TestShouldShowViolation(t *testing.T) {
	tests := []struct {
		operation string
		want      bool
	}{
		{operation: "network-outbound", want: true},
		{operation: "file-read-data", want: true},
		{operation: "file-write-create", want: true},
		{operation: "mach-lookup", want: true},
		{operation: "mach-register", want: true},
		{operation: "file-ioctl", want: false},
	}

	for _, tt := range tests {
		if got := shouldShowViolation(tt.operation); got != tt.want {
			t.Fatalf("shouldShowViolation(%q) = %v, want %v", tt.operation, got, tt.want)
		}
	}
}

func TestParseViolation_ShowsMachViolations(t *testing.T) {
	line := `2026-04-14 12:00:00.000000+0000 0x0 Default 0x0 0 0 kernel: Sandbox: Chromium(123) deny(1) mach-register org.chromium.Chromium.MachPortRendezvousServer`

	got := parseViolation(line)
	if got == "" {
		t.Fatal("expected mach-register violation to be shown")
	}
	if !strings.Contains(got, "mach-register") {
		t.Fatalf("expected output to mention mach-register, got %q", got)
	}
	if !strings.Contains(got, "org.chromium.Chromium.MachPortRendezvousServer") {
		t.Fatalf("expected output to mention denied service, got %q", got)
	}
}
