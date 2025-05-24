BINARY_NAME=plugin
BINARY_PATH=./bin/$(BINARY_NAME)
MAIN_GO_PATH=./main.go

# Default image name - can be overridden
IMAGE_NAME?=segator/vcluster-nodeport-plugin
IMAGE_TAG?=latest

# Go parameters
GO=go
GOFLAGS=-v
LDFLAGS="-s -w" # Strip debug symbols, reduce size

.PHONY: all
all: build

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build          Build the Go binary"
	@echo "  clean          Remove build artifacts"
	@echo "  docker-build   Build the Docker image"
	@echo "  docker-push    Push the Docker image to registry"
	@echo "  help           Show this help message"

.PHONY: build
build: go-build

.PHONY: go-build
go-build:
	@echo "Building Go binary..."
	@mkdir -p ./bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags=$(LDFLAGS) -o $(BINARY_PATH) $(MAIN_GO_PATH)
	@echo "Build complete: $(BINARY_PATH)"

.PHONY: docker-build
docker-build: go-build
	@echo "Building Docker image $(IMAGE_NAME):$(IMAGE_TAG)..."
	@docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .
	@echo "Docker image build complete: $(IMAGE_NAME):$(IMAGE_TAG)"


.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_PATH)
	@rm -rf ./bin
	@echo "Clean complete."

.PHONY: push
push: build
	@echo "Pushing Docker image $(IMAGE_NAME):$(IMAGE_TAG)..."
	@docker push $(IMAGE_NAME):$(IMAGE_TAG)
	@echo "Docker image push complete."
