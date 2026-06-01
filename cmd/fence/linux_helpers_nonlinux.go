//go:build !linux

package main

import (
	"os"

	"github.com/fencesandbox/fence/internal/fencelog"
)

func runLinuxInternalHelperMode(args []string) bool {
	if len(args) < 2 {
		return false
	}
	switch args[1] {
	case "--linux-argv-exec-run", "--linux-argv-exec-shim":
		fencelog.Printf("[fence:linux] %s is only available on Linux\n", args[1])
		os.Exit(1)
	}
	return false
}
