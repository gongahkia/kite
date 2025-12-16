.PHONY: help install test lint format clean build docker

help:
	@echo "Available commands:"
	@echo "  make install  - Install Python dependencies"
	@echo "  make test     - Run tests"
	@echo "  make lint     - Run linters"
	@echo "  make format   - Format code"
	@echo "  make clean    - Clean build artifacts"
	@echo "  make build    - Build Nim binary"
	@echo "  make docker   - Build Docker image"

install:
	pip install -e ".[dev]"

test:
	pytest tests/ -v --cov=kite

lint:
	flake8 kite/ tests/
	mypy kite/

format:
	black kite/ tests/
	isort kite/ tests/

clean:
	rm -rf build/ dist/ *.egg-info htmlcov/ .coverage .pytest_cache/
	find . -type d -name __pycache__ -exec rm -rf {} + 2>/dev/null || true
	find . -type f -name "*.pyc" -delete

build:
	nimble build -d:release

docker:
	docker build -t kite:latest .
