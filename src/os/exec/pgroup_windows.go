// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package exec

import (
	"internal/syscall/windows"
	"os"
	"slices"
	"syscall"
)

// sysProcAttr returns the SysProcAttr to start c with.
func (c *Cmd) sysProcAttr() (*syscall.SysProcAttr, error) {
	if !c.ProcessGroup {
		return c.SysProcAttr, nil
	}
	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return nil, os.NewSyscallError("CreateJobObject", err)
	}
	c.job = uintptr(job)

	var attr syscall.SysProcAttr
	if c.SysProcAttr != nil {
		attr = *c.SysProcAttr
	}
	attr.Jobs = append(slices.Clip(attr.Jobs), job)
	return &attr, nil
}

func (c *Cmd) killProcessGroup() error {
	// Kill the leader first: once the job has terminated it,
	// TerminateProcess fails with ERROR_ACCESS_DENIED.
	err := c.Process.Kill()
	if jerr := windows.TerminateJobObject(syscall.Handle(c.job), 1); jerr != nil {
		return os.NewSyscallError("TerminateJobObject", jerr)
	}
	return err
}

func (c *Cmd) releaseProcessGroup() {
	if c.job != 0 {
		syscall.CloseHandle(syscall.Handle(c.job))
		c.job = 0
	}
}
