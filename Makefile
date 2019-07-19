#!/usr/bin/env make

# Version of the entire package. Do not forget to update this when it's time
# to bump the version.
VERSION = v0.1.0

# Build tag. Useful to distinguish between same-version builds, but from
# different commits.
BUILD = $(shell git rev-parse --short HEAD)

# Full version includes both semantic version and git ref if present.
ifeq (${BUILD},)
	FULL_VERSION = $(VERSION)
else
	FULL_VERSION = $(VERSION)-$(BUILD)
endif

GO ?= go

TARGET_DIR := target

HOSTOS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
HOSTARCH := $(shell uname -m)

GOOS ?= $(HOSTOS)
GOARCH ?= $(HOSTARCH)

# Set the execution extension for Windows.
ifeq (${GOOS},windows)
    EXE := .exe
endif

OS_ARCH := $(GOOS)_$(GOARCH)

LDFLAGS = -X main.AppVersion=$(FULL_VERSION)

build/homekit:
	$(GO) build -ldflags "$(LDFLAGS)" -o $(TARGET_DIR)/$(OS_ARCH)/homekit cmd/homekit.go

clean:
	rm -rf $(TARGET_DIR)
