// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !unix && !windows

package exec

import (
	"errors"
	"runtime"
	"syscall"
)

// sysProcAttr returns the SysProcAttr to start c with.
func (c *Cmd) sysProcAttr() (*syscall.SysProcAttr, error) {
	if c.ProcessGroup {
		return nil, errors.New("exec: ProcessGroup is not supported on " + runtime.GOOS)
	}
	return c.SysProcAttr, nil
}

// killProcessGroup is unreachable: Start fails if ProcessGroup is set.
func (c *Cmd) killProcessGroup() error {
	return c.Process.Kill()
}

func (c *Cmd) releaseProcessGroup() {}
