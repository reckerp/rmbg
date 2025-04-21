# Configuration
BINARY_NAME=rmbg
BINDIR=/usr/local/bin
GO=go
GOFLAGS=-ldflags="-s -w" # Strip debug information for smaller binary

.PHONY: all build clean install uninstall

all: build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	$(GO) mod tidy
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) .
	@echo "Build complete."

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	@echo "Clean complete."

# Install the binary to BINDIR
install: build
	@echo "Installing $(BINARY_NAME) to $(BINDIR)..."
	@if [ -f "$(BINDIR)/$(BINARY_NAME)" ]; then \
		echo "Removing existing installation..."; \
		rm -f "$(BINDIR)/$(BINARY_NAME)"; \
	fi
	@mkdir -p $(BINDIR)
	cp $(BINARY_NAME) $(BINDIR)/
	@echo "Installation complete."

# Uninstall the binary
uninstall:
	@echo "Uninstalling $(BINARY_NAME) from $(BINDIR)..."
	rm -f "$(BINDIR)/$(BINARY_NAME)"
	@echo "Uninstallation complete."

