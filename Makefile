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

build/homekit:
	$(GO) build -o $(TARGET_DIR)/$(OS_ARCH)/homekit cmd/homekit.go

clean:
	rm -rf $(TARGET_DIR)
