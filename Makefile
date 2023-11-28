# Makefile for building and running the Go chat application

# Application name
APP_NAME := golang-chat

# Go related variables
GOBASE := $(shell pwd)
GOBIN := $(GOBASE)/bin
GOFILES := $(wildcard *.go)

# Build targets
.PHONY: all build-linux build-osx run clean

all: run

build: main.go
	GOOS=linux GOARCH=amd64 go build -ldflags '-s -w' -o $@
	
build-osx: main.go
	GOOS=darwin GOARCH=amd64 go build -ldflags '-s -w' -o $@

build-linux:
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 go build -o $(GOBIN)/$(APP_NAME)-linux $(GOFILES)

build-osx-m1:
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=arm64 go build -o $(GOBIN)/$(APP_NAME)-osx $(GOFILES)

run:
	@echo "Running application..."
	@go run $(GOFILES)

clean:
	@echo "Cleaning up..."
	@rm -rf $(GOBIN)