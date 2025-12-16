# Version 2 (Python)

Version 2 is the production-ready iteration with comprehensive observability, enterprise deployment support, and advanced legal research features.

- Language: Python 3.9+
- Architecture: Base scraper pattern with jurisdiction-specific implementations
- Scrapers: 17+ jurisdictions (US, UK, Canada, Australia, EU, Asia, International)
- Features: 
  - Citation extraction and network analysis
  - Legal concept tagging (200+ concepts)
  - Data validation and quality metrics
  - Ethical scraping compliance (robots.txt, rate limiting, attribution)
  - Jurisdiction metadata enrichment
- Observability: Prometheus metrics, structured logging (structlog), performance monitoring
- Testing: pytest with coverage, integration tests, e2e tests
- CI/CD: GitHub Actions, pre-commit hooks, security scanning
- Deployment: Docker, Kubernetes, docker-compose, Grafana dashboards
- Package: Published to PyPI, CLI and library usage

This folder archives v2 development; sources remain at repository root under `kite/` for backward compatibility.
