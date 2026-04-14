package main

import "testing"

func TestIsPureCDCommand(t *testing.T) {
	testCases := []struct {
		name    string
		command string
		want    bool
	}{
		{
			name:    "plain cd",
			command: "cd ../repo",
			want:    true,
		},
		{
			name:    "environment expansion",
			command: `cd "$HOME/tmp"`,
			want:    true,
		},
		{
			name:    "single quoted literal substitution syntax",
			command: `cd '$(pwd)'`,
			want:    true,
		},
		{
			name:    "command substitution",
			command: "cd $(pwd)",
			want:    false,
		},
		{
			name:    "command substitution in double quotes",
			command: `cd "$(pwd)"`,
			want:    false,
		},
		{
			name:    "backtick substitution",
			command: "cd `pwd`",
			want:    false,
		},
		{
			name:    "arithmetic expansion",
			command: "cd $((1 + 1))",
			want:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isPureCDCommand(tc.command); got != tc.want {
				t.Fatalf("expected %v, got %v for %q", tc.want, got, tc.command)
			}
		})
	}
}
