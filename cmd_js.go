//go:build js

package cmd

import (
	"os/exec"
	"syscall"
)

// 普通的cmd 客户端
func setAttr(cmd *exec.Cmd, detach bool) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// Setpgid:   true,
		// Pdeathsig: syscall.SIGTERM,
	}
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
