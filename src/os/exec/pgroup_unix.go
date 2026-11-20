// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build unix

package exec

import (
	"errors"
	"os"
	"syscall"
)

// sysProcAttr returns the SysProcAttr to start c with.
func (c *Cmd) sysProcAttr() (*syscall.SysProcAttr, error) {
	if !c.ProcessGroup {
		return c.SysProcAttr, nil
	}
	var attr syscall.SysProcAttr
	if c.SysProcAttr != nil {
		attr = *c.SysProcAttr
	}
	if (attr.Setpgid || attr.Foreground) && attr.Pgid != 0 {
		return nil, errors.New("exec: ProcessGroup is set, but SysProcAttr.Pgid joins an existing process group")
	}
	// Setsid already makes the child a process group leader.
	if !attr.Setsid {
		attr.Setpgid = true
	}
	return &attr, nil
}

func (c *Cmd) killProcessGroup() error {
	// Once Wait has reaped the leader, its pid may belong to another group.
	if err := c.Process.Signal(syscall.Signal(0)); err != nil {
		return err
	}
	err := syscall.Kill(-c.Process.Pid, syscall.SIGKILL)
	if err == syscall.ESRCH {
		return os.ErrProcessDone
	}
	return err
}

func (c *Cmd) releaseProcessGroup() {}
