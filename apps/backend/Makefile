BINARY_NAME = server
BUILD_DIR   = bin
GO_FILES    = $(shell find . -name '*.go' -not -path './vendor/*')

.PHONY: all build run test lint fmt vet clean docker-build docker-run tidy help

all: build

## build: compile the binary
build:
	go build -ldflags "-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/api/...

## run: run the server directly
run:
	go run ./cmd/api/...

## test: run all tests with race detector and coverage
test:
	go test -race -cover ./...

## lint: run golangci-lint
lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install it from https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run ./...

## fmt: format all Go source files
fmt:
	gofmt -w .

## vet: run go vet
vet:
	go vet ./...

## clean: remove build artifacts
clean:
	rm -rf $(BUILD_DIR)/

## docker-build: build the Docker image
docker-build:
	docker build -t portfolio-backend .

## docker-run: run the Docker container
docker-run:
	docker run --rm -p 8080:8080 portfolio-backend

## tidy: tidy go.mod and go.sum
tidy:
	go mod tidy

## help: display available make targets
help:
	@grep -E '^##' $(MAKEFILE_LIST) | sed 's/## //'
