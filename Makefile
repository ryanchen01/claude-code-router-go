# Go compiler
GO := go

# Output directory
OUT_DIR := ./bin

# Name of the executable
EXT :=

# Detect operating system
OS := $(shell uname -s)
ifeq ($(OS),Windows_NT)
    EXT := .exe
else
    ifeq ($(shell uname -o), Msys)
        EXT := .exe
	else
		EXT :=
    endif
endif

.PHONY: generate all build
all: build

build:
	@mkdir -p $(OUT_DIR)
	$(GO) build -o $(OUT_DIR)/claude-code-router$(EXT) ./cmd/main.go

generate:
	@go mod download
	@go generate ./...