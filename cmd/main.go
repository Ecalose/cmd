package main

import "github.com/gospider007/cmd"

// GOOS=windows GOARCH=amd64 go build -o ./cmd/cmd/watchdogWin.exe ./cmd/cmd/main.go  && GOOS=linux GOARCH=amd64 go build -o ./cmd/cmd/watchdogLinux ./cmd/cmd/main.go  && GOOS=darwin GOARCH=arm64 go build -o ./cmd/cmd/watchdogMac ./cmd/cmd/main.go
func main() {
	cmd.PrKill()
}
