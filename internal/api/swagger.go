package api

// Package api provides HTTP API handlers and routes
//
// @title Kite API
// @version 4.0.0
// @description Legal case law scraping and analysis API
// @description
// @description Kite v4 is an API-first backend service for scraping, storing, and analyzing
// @description legal case law from multiple jurisdictions worldwide.
// @description
// @description Features:
// @description - Multi-jurisdiction case law scraping (US, Canada, UK, Ireland, Australia, Hong Kong, etc.)
// @description - Citation extraction and network analysis
// @description - Legal concept tagging and taxonomy
// @description - Full-text search and filtering
// @description - RESTful API with OpenAPI documentation
// @description - Job queue for distributed scraping
// @description - Prometheus metrics and observability
//
// @contact.name Kite API Support
// @contact.url https://github.com/gongahkia/kite
// @contact.email support@kite-api.example.com
//
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
//
// @host localhost:8080
// @BasePath /api/v1
//
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
// @description API key for authentication. Obtain from your account dashboard.
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Bearer token. Format: "Bearer {token}"
//
// @tag.name Health
// @tag.description Health check and readiness endpoints
//
// @tag.name Cases
// @tag.description Case law retrieval and search
//
// @tag.name Judges
// @tag.description Judge information and statistics
//
// @tag.name Citations
// @tag.description Citation extraction and network analysis
//
// @tag.name Stats
// @tag.description System and database statistics
//
// @tag.name Scraper
// @tag.description Scraping job management
//
// @tag.name Search
// @tag.description Advanced search and query
//
// @x-logo {"url": "https://raw.githubusercontent.com/gongahkia/kite/main/docs/logo.png", "altText": "Kite Logo"}
