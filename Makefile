# Variables
APP_NAME := terium_node
CMD_DIR := ./cmd/app
BIN_DIR := ./bin
SRC_FILE := $(CMD_DIR)/main.go
OUTPUT_FILE := $(BIN_DIR)/$(APP_NAME)
GO := go

# Default target
.PHONY: all
all: build

# Build the application
.PHONY: build
build: $(OUTPUT_FILE)

$(OUTPUT_FILE): $(SRC_FILE)
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $@ $^
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
