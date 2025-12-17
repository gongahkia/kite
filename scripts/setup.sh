#!/bin/bash
set -e

echo "============================================"
echo "Kite v4 Development Environment Setup"
echo "============================================"
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed. Please install Go 1.22+ from https://go.dev/dl/"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo "✓ Go version: $GO_VERSION"

# Install Go dependencies
echo ""
echo "Installing Go dependencies..."
go mod download
go mod verify
echo "✓ Go dependencies installed"

# Install development tools
echo ""
echo "Installing development tools..."

# golangci-lint for linting
if ! command -v golangci-lint &> /dev/null; then
    echo "  Installing golangci-lint..."
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    echo "  ✓ golangci-lint installed"
else
    echo "  ✓ golangci-lint already installed"
fi

# air for hot reload during development
if ! command -v air &> /dev/null; then
    echo "  Installing air (hot reload)..."
    go install github.com/cosmtrek/air@latest
    echo "  ✓ air installed"
else
    echo "  ✓ air already installed"
fi

# mockgen for generating mocks
if ! command -v mockgen &> /dev/null; then
    echo "  Installing mockgen..."
    go install github.com/golang/mock/mockgen@latest
    echo "  ✓ mockgen installed"
else
    echo "  ✓ mockgen already installed"
fi

# Create local config file if it doesn't exist
echo ""
echo "Setting up configuration..."
if [ ! -f "configs/local.yaml" ]; then
    cp configs/default.yaml configs/local.yaml
    echo "✓ Created configs/local.yaml from default"
else
    echo "✓ configs/local.yaml already exists"
fi

# Create data directory for SQLite
echo ""
echo "Creating data directories..."
mkdir -p data
mkdir -p logs
echo "✓ Data directories created"

# Generate mocks for testing
echo ""
echo "Generating test mocks..."
bash scripts/generate-mocks.sh
echo "✓ Test mocks generated"

echo ""
echo "============================================"
echo "Setup Complete!"
echo "============================================"
echo ""
echo "Next steps:"
echo "  1. Review and update configs/local.yaml if needed"
echo "  2. Run 'make run' to start the API server"
echo "  3. Run 'make test' to run the test suite"
echo "  4. Run 'air' for hot reload during development"
echo ""
echo "Useful commands:"
echo "  make help        - Show all available commands"
echo "  make lint        - Run linter"
echo "  make test        - Run tests"
echo "  make coverage    - Generate test coverage report"
echo "  make docker-build - Build Docker image"
echo ""
