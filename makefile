BINARY_NAME=plugin
BINARY_PATH=./bin/$(BINARY_NAME)
MAIN_GO_PATH=./cmd/plugin/main.go

# Default image name - can be overridden
IMAGE_NAME?=your-ghcr-username/your-repo-name/vcluster-nodeport-patcher
IMAGE_TAG?=latest

# Go parameters
GO=go
GOFLAGS=-v
LDFLAGS="-s -w" # Strip debug symbols, reduce size

.PHONY: all build clean docker-build docker-push help

all: build

help:
	@echo "Available targets:"
	@echo "  build          Build the Go binary"
	@echo "  clean          Remove build artifacts"
	@echo "  docker-build   Build the Docker image"
	@echo "  docker-push    Push the Docker image to registry"
	@echo "  help           Show this help message"

build:
	@echo "Building Go binary..."
	@mkdir -p ./bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags=$(LDFLAGS) -o $(BINARY_PATH) $(MAIN_GO_PATH)
	@echo "Build complete: $(BINARY_PATH)"

clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_PATH)
	@rm -rf ./bin
	@echo "Clean complete."

docker-build: build
	@echo "Building Docker image $(IMAGE_NAME):$(IMAGE_TAG)..."
	@docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .
	@echo "Docker image build complete: $(IMAGE_NAME):$(IMAGE_TAG)"

docker-push:
	@echo "Pushing Docker image $(IMAGE_NAME):$(IMAGE_TAG)..."
	@docker push $(IMAGE_NAME):$(IMAGE_TAG)
	@echo "Docker image push complete."
