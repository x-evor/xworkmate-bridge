.DEFAULT_GOAL := help

SHELL := /bin/bash
GO ?= go
OUTPUT_DIR ?= $(CURDIR)/build/bin
OUTPUT_PATH ?= $(OUTPUT_DIR)/xworkmate-go-core

.PHONY: help test build clean

help:
	@printf "%-12s %s\n" "test" "Run Go tests"
	@printf "%-12s %s\n" "build" "Build xworkmate-go-core helper"
	@printf "%-12s %s\n" "clean" "Remove build output"

test:
	$(GO) test ./...

build:
	bash scripts/build-helper.sh

clean:
	rm -rf build
