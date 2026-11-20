// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package exec_test

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"runtime"
	"testing"
	"time"
)

// startHangGroup starts the hang helper with ProcessGroup set.
// The helper leaves a grandchild that holds the write end of the returned
// pipe, so reading the pipe to EOF shows that the grandchild is gone.
func startHangGroup(t *testing.T, ctx context.Context, setup func(*exec.Cmd)) (*exec.Cmd, *os.File) {
	t.Helper()
	if runtime.GOOS == "plan9" {
		t.Skip("ProcessGroup is not supported on plan9")
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { r.Close() })

	cmd := helperCommandContext(t, ctx, "hang", "10m", "-subsleep=10m")
	cmd.ProcessGroup = true
	cmd.Stderr = w
	if setup != nil {
		setup(cmd)
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	err = cmd.Start()
	w.Close()
	if err != nil {
		t.Fatal(err)
	}

	// The helper closes stdout once the grandchild is running.
	if _, err := io.Copy(io.Discard, out); err != nil {
		cmd.Process.Kill()
		cmd.Wait()
		t.Fatal(err)
	}
	return cmd, r
}

func waitGroupKilled(t *testing.T, cmd *exec.Cmd, stderr *os.File) {
	t.Helper()
	err := cmd.Wait()
	if ee, ok := errors.AsType[*exec.ExitError](err); !ok {
		t.Errorf("Wait error = %v; want %T", err, ee)
	}

	b, err := io.ReadAll(stderr)
	t.Logf("stderr:\n%s", b)
	if err != nil {
		t.Error(err)
	}
}

func TestProcessGroup(t *testing.T) {
	t.Parallel()

	t.Run("Cancel", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		cmd, stderr := startHangGroup(t, ctx, nil)
		cancel()
		waitGroupKilled(t, cmd, stderr)
	})

	// A Cancel that does nothing leaves the kill to WaitDelay.
	t.Run("WaitDelay", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		cmd, stderr := startHangGroup(t, ctx, func(cmd *exec.Cmd) {
			cmd.Cancel = func() error { return nil }
			cmd.WaitDelay = 10 * time.Millisecond
		})
		cancel()
		waitGroupKilled(t, cmd, stderr)
	})
}
