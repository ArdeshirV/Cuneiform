# ==============================================================================
# Fyne Makefile — Cross-platform build system
# Requirements: GNU Make, Go, Fyne CLI
# ==============================================================================

.DEFAULT_GOAL := help
.DELETE_ON_ERROR:
.SUFFIXES:

empty :=
space := $(empty) $(empty)
.RECIPEPREFIX := $(space)

# Application
APP       ?= $(notdir $(CURDIR))
MAIN      ?= .
FYNE      ?= fyne
GO        ?= go
OUT       ?= build
RELEASE   ?= true

# Metadata
APP_ID    ?=
ICON      ?=
TAGS      ?=
LDFLAGS   ?=
GOFLAGS   ?=
FYNEFLAGS ?=

# Host platform
HOST_OS   := $(shell $(GO) env GOOS)
HOST_ARCH := $(shell $(GO) env GOARCH)

# Common arguments
RELEASE_FLAG = $(if $(filter true,$(RELEASE)),--release,)
META_FLAGS   = $(if $(APP_ID),-app-id "$(APP_ID)",) \
               $(if $(ICON),-icon "$(ICON)",)

.PHONY: help check run build package test fmt vet tidy clean \
        linux linux-amd64 linux-arm64 linux-arm linux-386 \
        windows windows-amd64 windows-arm64 windows-386 \
        darwin darwin-amd64 darwin-arm64 \
        android android-arm64 android-arm android-amd64 android-386 \
        ios iossimulator wasm freebsd-amd64 freebsd-arm64

# Help
help:
        @echo "Fyne cross-platform build system"
        @echo ""
        @echo "Development: make run | build | test | fmt | vet | tidy | clean"
        @echo "Packaging:   make package OS=linux ARCH=amd64"
        @echo "Platforms:   linux windows darwin android ios iossimulator wasm"
        @echo ""
        @echo "Options: APP MAIN OUT RELEASE APP_ID ICON TAGS LDFLAGS GOFLAGS FYNEFLAGS"

# Verify tools
check:
        @$(GO) version
        @$(FYNE) version

# Development
run:
        $(GO) run $(GOFLAGS) $(if $(TAGS),-tags "$(TAGS)",) $(MAIN)

build:
        mkdir -p "$(OUT)"
        $(GO) build $(GOFLAGS) $(if $(TAGS),-tags "$(TAGS)",) \
                $(if $(LDFLAGS),-ldflags "$(LDFLAGS)",) -o "$(OUT)/$(APP)" $(MAIN)

test:
        $(GO) test $(GOFLAGS) ./...

fmt:
        $(GO) fmt ./...

vet:
        $(GO) vet $(GOFLAGS) ./...

tidy:
        $(GO) mod tidy

# General Fyne package target
package:
        GOOS="$(OS)" GOARCH="$(ARCH)"
        $(FYNE) package -os "$(OS)" $(RELEASE_FLAG) \
                $(META_FLAGS) $(FYNEFLAGS)

# Linux
linux: linux-amd64

linux-amd64:
        $(MAKE) package OS=linux ARCH=amd64

linux-arm64:
        $(MAKE) package OS=linux ARCH=arm64

linux-arm:
        $(MAKE) package OS=linux ARCH=arm

linux-386:
        $(MAKE) package OS=linux ARCH=386

# Windows
WINDOWS_AMD64_CC ?= x86_64-w64-mingw32-gcc

windows: windows-amd64

windows-amd64:
        GOOS=windows GOARCH=amd64 CC="$(WINDOWS_AMD64_CC)" $(FYNE) package -os windows $(RELEASE_FLAG) $(META_FLAGS) $(FYNEFLAGS)

windows-arm64:
        GOOS=windows GOARCH=arm64 $(FYNE) package -os windows $(RELEASE_FLAG) $(META_FLAGS) $(FYNEFLAGS)

windows-386:
        GOOS=windows GOARCH=386 $(FYNE) package -os windows $(RELEASE_FLAG) $(META_FLAGS) $(FYNEFLAGS)

# macOS
darwin: darwin-arm64

darwin-arm64:
        GOOS=darwin GOARCH=arm64 CC=oa64-clang CXX=oa64-clang++
        $(FYNE) package -os darwin $(RELEASE_FLAG) $(META_FLAGS) $(FYNEFLAGS)

darwin-amd64:
        GOOS=darwin GOARCH=amd64 CC=o64-clang CXX=o64-clang++
        $(FYNE) package -os darwin $(RELEASE_FLAG) $(META_FLAGS) $(FYNEFLAGS)

# Android
android:
        $(FYNE) package -os android $(RELEASE_FLAG) $(META_FLAGS) $(FYNEFLAGS)

android-arm64:
        $(FYNE) package -os android/arm64 $(RELEASE_FLAG) $(META_FLAGS) $(FYNEFLAGS)

android-arm:
        $(FYNE) package -os android/arm $(RELEASE_FLAG) $(META_FLAGS) $(FYNEFLAGS)

android-amd64:
        $(FYNE) package -os android/amd64 $(RELEASE_FLAG) $(META_FLAGS) $(FYNEFLAGS)

android-386:
        $(FYNE) package -os android/386 $(RELEASE_FLAG) $(META_FLAGS) $(FYNEFLAGS)

# Apple mobile
ios:
        $(FYNE) package -os ios $(RELEASE_FLAG) $(META_FLAGS) $(FYNEFLAGS)

iossimulator:
        $(FYNE) package -os iossimulator $(RELEASE_FLAG) $(META_FLAGS) $(FYNEFLAGS)

# WebAssembly
wasm:
        $(FYNE) package -os wasm $(RELEASE_FLAG) $(META_FLAGS) $(FYNEFLAGS)

# FreeBSD — requires a suitable cross-compilation toolchain
freebsd-amd64:
        GOOS=freebsd GOARCH=amd64 $(GO) build -o "$(OUT)/$(APP)" $(MAIN)

freebsd-arm64:
        GOOS=freebsd GOARCH=arm64 $(GO) build -o "$(OUT)/$(APP)" $(MAIN)

# Cleanup
clean:
        $(GO) clean
        $(RM) -r "$(OUT)"
