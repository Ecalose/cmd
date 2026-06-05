package cmd

import (
	"context"
	"errors"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func processExists(pid int) bool {
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// 信号0仅用于检测
	err = p.Signal(syscall.Signal(0))
	return err == nil || errors.Is(err, syscall.EPERM)
}
func runChrome(ctx context.Context) (*Client, error) {
	name := os.Args[2]
	args := os.Args[3:]
	var userDir string
	for _, arg := range args {
		v, ok := strings.CutPrefix(arg, "--user-data-dir=")
		if ok {
			userDir = v
		}
	}
	option := ClientOption{Name: name, Args: args}
	if userDir != "" {
		option.CloseCallBack = func() {
			for range 10 {
				if os.RemoveAll(userDir) == nil {
					return
				}
				time.Sleep(time.Millisecond * 300)
			}
		}
	}
	return NewClient(ctx, option)
}
func PrKill() {
	if len(os.Args) < 3 {
		return
	}
	ppid, err := strconv.Atoi(os.Args[1])
	if err != nil || !processExists(ppid) {
		return
	}
	ctx, cnl := context.WithCancel(context.TODO())
	defer cnl()
	cli, err := runChrome(ctx)
	if err != nil {
		return
	}
	defer cli.Close()
	go cli.Run()
pe:
	for processExists(ppid) {
		select {
		case <-cli.Ctx().Done():
			break pe
		case <-time.After(time.Second * 2):
		}
	}
}
