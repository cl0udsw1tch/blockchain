# Variables
APP_NAME := tiereum
CMD_DIR := ./cmd/app
BIN_DIR := ./bin
MAIN_FILE := $(CMD_DIR)/main.go
OUTPUT_FILE := $(BIN_DIR)/$(APP_NAME)
GO := go
GOFILES := $(shell find . -name '*.go')

TRM_ROOT_DIR := $(HOME)/tiereum/client
SHELL := /bin/bash

.PHONY: all
all: build
 
.PHONY: build
build: $(OUTPUT_FILE)

$(OUTPUT_FILE): $(GOFILES)
	@echo "Building $(APP_NAME)..."
	@if [ -d $(BIN_DIR) ]; then \
	mkdir -p $(BIN_DIR); \
	else echo "Wont be recreating bin directory"; \
	fi;
	$(GO) build -o ./bin/$(APP_NAME) ./cmd/app
	@echo "Build completed: $@"

.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(OUTPUT_FILE)
	@echo "Clean completed."

.PHONY: setup


setup:
	@if [ ! -d "$(TRM_ROOT_DIR)" ]; then \
		mkdir -p "$(TRM_ROOT_DIR)"; \
	fi;
	@if [ ! -d "$(TRM_ROOT_DIR)/wallets" ]; then \
		mkdir -p "$(TRM_ROOT_DIR)/wallets"; \
	fi;
	@if [ ! -f "$(TRM_ROOT_DIR)/networkConfig.json" ]; then \
		touch "$(TRM_ROOT_DIR)/networkConfig.json"; \
		echo -e "{\n\"clientPort\": 7033,\n\"nodeTable\": [\n\"http://34.16.104.117:8033\"\n]\n}" >> "$(TRM_ROOT_DIR)/networkConfig.json"; \
	fi;

