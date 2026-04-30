package sandbox

import (
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

// startSleepingChild spawns a long-lived child whose Process can be
// observed/signaled, and registers a cleanup to reap it. The child
// runs in its own pgrp so PgrpBroadcast tests are meaningful.
func startSleepingChild(t *testing.T) *exec.Cmd {
	t.Helper()
	cmd := exec.Command("sleep", "30")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pgid: 0}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start child: %v", err)
	}
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	})
	return cmd
}

func TestSignalForwarder_StopIsIdempotent(t *testing.T) {
	cmd := startSleepingChild(t)
	stop := (&SignalForwarder{Cmd: cmd}).Start()
	stop()
	stop()
	stop()
}

func TestSignalForwarder_SIGWINCHDoesNotEscalate(t *testing.T) {
	cmd := startSleepingChild(t)

	var escalates atomic.Int32
	f := &SignalForwarder{
		Cmd:        cmd,
		OnEscalate: func() { escalates.Add(1) },
	}

	sigChan := make(chan os.Signal, 4)
	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		f.run(sigChan, done)
	}()

	for i := 0; i < 3; i++ {
		sigChan <- syscall.SIGWINCH
	}
	// Give run() a moment to drain the channel.
	time.Sleep(20 * time.Millisecond)
	close(done)
	wg.Wait()

	if got := escalates.Load(); got != 0 {
		t.Fatalf("OnEscalate fired %d times for SIGWINCH-only stream; want 0", got)
	}
}

func TestSignalForwarder_FirstSignalDoesNotEscalate(t *testing.T) {
	cmd := startSleepingChild(t)

	var escalates atomic.Int32
	f := &SignalForwarder{
		Cmd:        cmd,
		OnEscalate: func() { escalates.Add(1) },
	}

	sigChan := make(chan os.Signal, 1)
	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		f.run(sigChan, done)
	}()

	sigChan <- syscall.SIGINT
	time.Sleep(20 * time.Millisecond)
	close(done)
	wg.Wait()

	if got := escalates.Load(); got != 0 {
		t.Fatalf("OnEscalate fired %d times after 1st signal; want 0", got)
	}
}

func TestSignalForwarder_SecondSignalEscalates(t *testing.T) {
	cmd := startSleepingChild(t)

	var escalates atomic.Int32
	f := &SignalForwarder{
		Cmd:        cmd,
		OnEscalate: func() { escalates.Add(1) },
	}

	sigChan := make(chan os.Signal, 2)
	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		f.run(sigChan, done)
	}()

	sigChan <- syscall.SIGINT
	sigChan <- syscall.SIGINT
	// Wait for child to die so we know run() processed both signals.
	waitErrCh := make(chan error, 1)
	go func() { waitErrCh <- cmd.Wait() }()
	select {
	case <-waitErrCh:
	case <-time.After(2 * time.Second):
		t.Fatal("child did not die after 2nd signal escalation")
	}
	close(done)
	wg.Wait()

	if got := escalates.Load(); got != 1 {
		t.Fatalf("OnEscalate fired %d times after 2nd signal; want 1", got)
	}
}

func TestSignalForwarder_SIGHUPParticipatesInEscalation(t *testing.T) {
	cmd := startSleepingChild(t)

	var escalates atomic.Int32
	f := &SignalForwarder{
		Cmd:        cmd,
		OnEscalate: func() { escalates.Add(1) },
	}

	sigChan := make(chan os.Signal, 2)
	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		f.run(sigChan, done)
	}()

	sigChan <- syscall.SIGHUP // 1st: forwarded
	sigChan <- syscall.SIGHUP // 2nd: escalation
	waitErrCh := make(chan error, 1)
	go func() { waitErrCh <- cmd.Wait() }()
	select {
	case <-waitErrCh:
	case <-time.After(2 * time.Second):
		t.Fatal("child did not die after 2nd SIGHUP")
	}
	close(done)
	wg.Wait()

	if got := escalates.Load(); got != 1 {
		t.Fatalf("OnEscalate after SIGHUP escalation fired %d times; want 1", got)
	}
}

func TestSignalForwarder_NilOnEscalateIsSafe(t *testing.T) {
	cmd := startSleepingChild(t)
	f := &SignalForwarder{Cmd: cmd, OnEscalate: nil}

	sigChan := make(chan os.Signal, 2)
	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		f.run(sigChan, done)
	}()

	sigChan <- syscall.SIGINT
	sigChan <- syscall.SIGINT
	waitErrCh := make(chan error, 1)
	go func() { waitErrCh <- cmd.Wait() }()
	select {
	case <-waitErrCh:
	case <-time.After(2 * time.Second):
		t.Fatal("child did not die with nil OnEscalate")
	}
	close(done)
	wg.Wait()
}

func TestSignalForwarder_NilProcessIgnored(t *testing.T) {
	// f.Cmd.Process is nil before Start(); run() must not panic.
	f := &SignalForwarder{Cmd: &exec.Cmd{}}

	sigChan := make(chan os.Signal, 2)
	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		f.run(sigChan, done)
	}()

	sigChan <- syscall.SIGINT
	sigChan <- syscall.SIGTERM
	time.Sleep(20 * time.Millisecond)
	close(done)
	wg.Wait()
}

func TestKillProcessGroup_NonPositiveLeaderIsNoOp(t *testing.T) {
	if err := killProcessGroup(0, syscall.SIGTERM); err != nil {
		t.Fatalf("killProcessGroup(0, ...) = %v; want nil", err)
	}
	if err := killProcessGroup(-1, syscall.SIGTERM); err != nil {
		t.Fatalf("killProcessGroup(-1, ...) = %v; want nil", err)
	}
}
