.PHONY: all run clean fmt vet tidy build

APP_NAME := blog-app
BUILD_DIR := ./bin
MAIN_GO_FILE := ./cmd/app/main.go
TEST_FILES_DIR := ./tests/...

all: fmt vet tidy run

fmt:
	@echo "Formatting Go code..."
	@go fmt ./...
	@echo "Formatting complete."

vet:
	@echo "Running static analysis (go vet)..."
	@go vet ./...
	@echo "Analysis complete."

tidy:
	@echo "Cleaning and synchronizing Go modules (go mod tidy)"
	@go mod tidy
	@echo "Modules synchronized."

build: fmt vet tidy
	@echo "Building app..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME) -tags "linux" $(MAIN_GO_FILE)
	@echo "Build complete."

run: build
	@echo "Running Go application..."
	@$(BUILD_DIR)/$(APP_NAME)
	@echo "Application finished."

clean:
	@echo "Cleaning build directory..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete."
