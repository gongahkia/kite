#!/bin/bash
set -e

echo "============================================"
echo "Generating Test Mocks for Kite v4"
echo "============================================"
echo ""

# Check if mockgen is installed
if ! command -v mockgen &> /dev/null; then
    echo "Error: mockgen not installed"
    echo "Installing mockgen..."
    go install github.com/golang/mock/mockgen@latest
fi

# Create mocks directory
mkdir -p test/mocks

echo "Generating mocks..."
echo ""

# Generate mock for Queue interface
echo "  Generating mock for Queue interface..."
mockgen -source=internal/queue/interface.go \
    -destination=test/mocks/queue_mock.go \
    -package=mocks \
    Queue

# Generate mock for Storage interface
echo "  Generating mock for Storage interface..."
mockgen -source=internal/storage/interface.go \
    -destination=test/mocks/storage_mock.go \
    -package=mocks \
    Storage

# Generate mock for Scraper interface
echo "  Generating mock for Scraper interface..."
mockgen -source=internal/scraper/interface.go \
    -destination=test/mocks/scraper_mock.go \
    -package=mocks \
    Scraper

echo ""
echo "✓ All mocks generated successfully"
echo ""
echo "Mock files created in test/mocks/:"
echo "  - queue_mock.go"
echo "  - storage_mock.go"
echo "  - scraper_mock.go"
echo ""
