.PHONY: build build-all build-linux-amd64 build-darwin-amd64 build-darwin-arm64 build-linux-amd64-in-container krew-release release-assets clean install test test-linux

BINARY_NAME=kubectl-nks_ctx
VERSION?=v0.1.0
# Container runtime for test-linux / build-all-containers (podman or docker)
CONTAINER_RUNTIME ?= $(shell command -v podman 2>/dev/null || command -v docker 2>/dev/null || echo "podman")

build:
	@echo "Building $(BINARY_NAME)..."
	@CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BINARY_NAME) .

build-darwin-amd64:
	@echo "Building $(BINARY_NAME) for darwin/amd64..."
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $(BINARY_NAME)-darwin-amd64 .

build-darwin-arm64:
	@echo "Building $(BINARY_NAME) for darwin/arm64..."
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o $(BINARY_NAME)-darwin-arm64 .

build-linux-amd64:
	@echo "Building $(BINARY_NAME) for linux/amd64..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BINARY_NAME)-linux-amd64 .

# Cross-compile all platform binaries (run on Mac or Linux with Go installed)
build-all: build-darwin-amd64 build-darwin-arm64 build-linux-amd64
	@echo "All builds completed. Binaries:"
	@ls -la $(BINARY_NAME)-darwin-amd64 $(BINARY_NAME)-darwin-arm64 $(BINARY_NAME)-linux-amd64 2>/dev/null || true

# krew-release builds all binaries and creates tar.gz with binary + LICENSE for krew
krew-release: build-all
	@for platform in darwin-amd64 darwin-arm64 linux-amd64; do \
		mkdir -p dist/nks-ctx-$$platform && \
		cp $(BINARY_NAME)-$$platform dist/nks-ctx-$$platform/$(BINARY_NAME) && \
		cp LICENSE dist/nks-ctx-$$platform/ && \
		tar -C dist/nks-ctx-$$platform -czvf dist/$(BINARY_NAME)-$$platform.tar.gz . && \
		rm -rf dist/nks-ctx-$$platform; \
	done
	@echo "Release archives in dist/ (include LICENSE for krew)"

# Release assets for GitHub/Releases: source zip, source tar.gz, binary zips, checksums
release-assets: build-all
	@mkdir -p dist
	@echo "Creating source archives..."
	@git archive --format=zip --prefix=nks-ctx-$(VERSION)/ -o dist/sourcecode.zip HEAD
	@git archive --format=tar.gz --prefix=nks-ctx-$(VERSION)/ -o dist/sourcecode.tar.gz HEAD
	@echo "Creating binary zip files..."
	@for platform in darwin-amd64 darwin-arm64 linux-amd64; do \
		mkdir -p dist/nks-ctx-$$platform && \
		cp $(BINARY_NAME)-$$platform dist/nks-ctx-$$platform/$(BINARY_NAME) && \
		cp LICENSE dist/nks-ctx-$$platform/ && \
		(cd dist/nks-ctx-$$platform && zip -r ../$(BINARY_NAME)-$$platform.zip .) && \
		rm -rf dist/nks-ctx-$$platform; \
	done
	@echo "Creating checksums..."
	@cd dist && for f in *.zip *.tar.gz; do [ -f "$$f" ] && openssl dgst -sha256 -r "$$f"; done > checksums.txt
	@echo "Release assets in dist/:"
	@ls -la dist/

clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -f $(BINARY_NAME)-darwin-amd64
	@rm -f $(BINARY_NAME)-darwin-arm64
	@rm -f $(BINARY_NAME)-linux-amd64
	@rm -rf dist

install: build
	@echo "Installing $(BINARY_NAME)..."
	@cp $(BINARY_NAME) ~/.local/bin/$(BINARY_NAME) || cp $(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@chmod +x ~/.local/bin/$(BINARY_NAME) 2>/dev/null || chmod +x /usr/local/bin/$(BINARY_NAME) 2>/dev/null || true
	@echo "Installed! You can now use: kubectl nks_ctx"

test:
	@CGO_ENABLED=0 go test ./...

# Run tests inside a Linux container (uses podman or docker)
test-linux:
	@$(CONTAINER_RUNTIME) run --rm -v "$$(pwd):/app" -w /app golang:1.21-alpine sh -c "go test ./..."

# Build linux/amd64 binary inside a Linux container (uses podman or docker)
build-linux-amd64-in-container:
	@$(CONTAINER_RUNTIME) run --rm -v "$$(pwd):/app" -w /app -e CGO_ENABLED=0 -e GOOS=linux -e GOARCH=amd64 golang:1.21-alpine go build -ldflags="-s -w" -o $(BINARY_NAME)-linux-amd64 .
	@echo "Built $(BINARY_NAME)-linux-amd64 in Linux container"

test-verbose:
	@CGO_ENABLED=0 go test -v ./...

test-cover:
	@CGO_ENABLED=0 go test -cover ./...

test-cover-html:
	@CGO_ENABLED=0 go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

fmt:
	@go fmt ./...

vet:
	@go vet ./...

lint: fmt vet
	@echo "Linting completed"
