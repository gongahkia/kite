# Architecture

## Overview

Kite is a legal case law scraper with support for multiple jurisdictions.

## Components

- **Scrapers**: Individual scrapers for different legal databases
- **Utils**: Shared utilities for HTTP, parsing, caching
- **CLI**: Command-line interface
- **Metrics**: Prometheus metrics for monitoring

## Design Principles

- Modular design with pluggable scrapers
- Rate limiting and retry logic
- Comprehensive error handling
- Performance monitoring
