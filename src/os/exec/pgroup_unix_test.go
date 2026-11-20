// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build unix && !aix && !solaris

package exec_test

import (
	"fmt"
	"os"
	"syscall"
	"testing"
)

func init() {
	registerHelperCommand("pgrp", cmdPgrp)
}

func cmdPgrp(...string) {
	fmt.Println(os.Getpid(), syscall.Getpgrp())
}

func TestProcessGroupLeader(t *testing.T) {
	t.Parallel()

	attr := new(syscall.SysProcAttr)
	cmd := helperCommand(t, "pgrp")
	cmd.SysProcAttr = attr
	cmd.ProcessGroup = true
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}

	var pid, pgrp int
	if _, err := fmt.Sscan(string(out), &pid, &pgrp); err != nil {
		t.Fatalf("parsing %q: %v", out, err)
	}
	if pgrp != pid {
		t.Errorf("child process group = %d; want its own pid %d", pgrp, pid)
	}
	if attr.Setpgid {
		t.Errorf("Start modified cmd.SysProcAttr")
	}
}

func TestProcessGroupJoinExisting(t *testing.T) {
	t.Parallel()

	cmd := helperCommand(t, "pgrp")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pgid: syscall.Getpgrp()}
	cmd.ProcessGroup = true
	if err := cmd.Start(); err == nil {
		cmd.Wait()
		t.Fatal("Start succeeded; want error for Pgid of an existing process group")
	}
}
