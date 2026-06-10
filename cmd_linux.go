//go:build linux

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
		attr.Pdeathsig = syscall.SIGKILL
	}
	cmd.SysProcAttr = attr
}
func killProcess(cmd *exec.Cmd) {
	cmd.Process.Kill()
	syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) // Kill the process and its children
}
