CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o ./cmd/cmd/bin/watchdog-windows-amd64.exe ./cmd/cmd/main.go \
&& CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ./cmd/cmd/bin/watchdog-linux-amd64 ./cmd/cmd/main.go \
&& CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o ./cmd/cmd/bin/watchdog-mac-arm64 ./cmd/cmd/main.go