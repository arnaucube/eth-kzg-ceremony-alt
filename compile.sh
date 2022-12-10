#!/bin/sh

mkdir -p bin

cd cmd

echo "building linux binaries"
GOOS=linux GOARCH=amd64 go build -o ../bin/kzgceremony-amd64-linux *.go

echo "building windows binaries"
GOOS=windows GOARCH=amd64 go build -o ../bin/kzgceremony-amd64.exe *.go

echo "building macOS binaries"
GOOS=darwin GOARCH=amd64 go build -o ../bin/kzgceremony-amd64-darwin *.go
