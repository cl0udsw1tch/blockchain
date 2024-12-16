# Variables
APP_NAME := terium
CMD_DIR := ./cmd/app
BIN_DIR := ./bin
MAIN_FILE := $(CMD_DIR)/main.go
OUTPUT_FILE := $(BIN_DIR)/$(APP_NAME)
GO := go
GOFILES := $(shell find . -name '*.go')


# Default target
.PHONY: all
all: build
	
# Build the application
.PHONY: build
build: $(OUTPUT_FILE)

$(OUTPUT_FILE): $(GOFILES)
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BIN_DIR)
	$(GO) build -o ./bin/$(APP_NAME) ./cmd/app
	@echo "Build completed: $@"

# Run the application
.PHONY: run
run: build
	@echo "Running $(APP_NAME)..."
	$(OUTPUT_FILE)

# Clean the build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(OUTPUT_FILE)
	@echo "Clean completed."

# Format the Go code
.PHONY: fmt
fmt:
	@echo "Formatting Go code..."
	$(GO) fmt ./...
	@echo "Format completed."

# Install Go dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	$(GO) mod tidy
	@echo "Dependencies installed."

# Test the code
.PHONY: test
test:
	@echo "Running tests..."
	$(GO) test ./...
	@echo "Tests completed."

# Lint the code
.PHONY: lint
lint:
	@echo "Linting code..."
	$(GO) vet ./...
	@echo "Lint completed."

.PHONY: setup

setup: 
	@if [ ! -d $(HOME)/terium ]; then \
	mkdir $(HOME)/terium; \
	fi
	@if [ ! -d $(HOME)/terium/.data ]; then \
	mkdir $(HOME)/terium/.data; \
	fi
	@if [ ! -d $(HOME)/terium/.tmp ]; then \
	mkdir $(HOME)/terium/.tmp; \
	fi
	@if [ ! -d $(HOME)/terium/wallets ]; then \
	mkdir $(HOME)/terium/wallets; \
	fi
	@if [ ! -f $(HOME)/terium/config.json ]; then \
	echo '{}' > $(HOME)/terium/config.json; \
	fi
	@if [ ! -d $(HOME)/terium/.data/index ]; then \
	mkdir $(HOME)/terium/.data/index; \
	fi
