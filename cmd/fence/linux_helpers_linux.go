//go:build linux

package main

import (
	"os"

	"github.com/fencesandbox/fence/internal/fencelog"
	"github.com/fencesandbox/fence/internal/sandbox"
)

func runLinuxInternalHelperMode(args []string) bool {
	if len(args) >= 2 && args[1] == "--linux-argv-exec-run" {
		exitCode, err := sandbox.RunLinuxArgvExecRunnerFromEnv()
		if err != nil {
			fencelog.Printf("[fence:linux] %v\n", err)
		}
		os.Exit(exitCode)
	}
	if len(args) >= 2 && args[1] == "--linux-argv-exec-shim" {
		exitCode, err := sandbox.RunLinuxArgvExecShim(args[2:])
		if err != nil {
			fencelog.Printf("[fence:linux] %v\n", err)
		}
		os.Exit(exitCode)
	}
	return false
}
