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
func killProcess(cmd *exec.Cmd) {
	cmd.Process.Kill()
	syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) // Kill the process and its children
}
