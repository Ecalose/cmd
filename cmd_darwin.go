//go:build darwin

package cmd

import (
	"os/exec"
	"syscall"
)

// 普通的cmd 客户端
func setAttr(cmd *exec.Cmd, detach bool) {
	attr := &syscall.SysProcAttr{}
	if detach {
		attr.Setsid = true
	} else {
		attr.Setpgid = true
	}
	cmd.SysProcAttr = attr
}
func killProcess(cmd *exec.Cmd, detach bool) {
	if detach {
		cmd.Process.Signal(syscall.SIGTERM)
		syscall.Kill(cmd.Process.Pid, syscall.SIGTERM)
	} else {
		cmd.Process.Signal(syscall.SIGKILL)
		syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) // Kill the process and its children
	}
}
