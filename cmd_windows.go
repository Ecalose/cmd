//go:build windows

package cmd

import (
	"os/exec"
	"syscall"
)

// 普通的cmd 客户端
func setAttr(cmd *exec.Cmd, detach bool) {
	cmd.SysProcAttr = &syscall.SysProcAttr{}
}
func killProcess(cmd *exec.Cmd, detach bool) {
	if detach {
		cmd.Process.Signal(syscall.SIGTERM)
	} else {
		cmd.Process.Signal(syscall.SIGKILL)
	}
}
