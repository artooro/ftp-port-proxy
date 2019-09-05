#!/usr/bin/env zsh

GOOS=windows GOARCH=amd64 go build -o build/windows-amd64/ftp-port-proxy.exe
GOOS=freebsd GOARCH=amd64 go build -o build/freebsd-amd64/ftp-port-proxy
