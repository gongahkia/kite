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

## Usage

### Setup

```console
$ pip install kite==2.0.0
```

Or from source:

```console
$ cd archive/v2-python
$ pip install -e .
```

### Production Deployment

#### Docker

```console
$ cd archive/v2-python
$ docker-compose up -d
```

#### Kubernetes

```console
$ kubectl apply -f k8s/namespace.yaml
$ kubectl apply -f k8s/configmap.yaml
$ kubectl apply -f k8s/deployment.yaml
$ kubectl apply -f k8s/service.yaml
$ kubectl apply -f k8s/hpa.yaml
```

### V2-Specific Features

#### Citation Extraction and Network Analysis

```python
from kite import CourtListenerScraper
from kite.utils.citation_extractor import extract_citations
from kite.utils.citation_network import build_citation_graph

with CourtListenerScraper() as scraper:
    case = scraper.get_case_by_id("1234567")
    
    # Extract citations
    citations = extract_citations(case.text)
    print(f"Found {len(citations)} citations")
    
    # Build citation network
    graph = build_citation_graph([case])
    print(f"Graph has {graph.number_of_nodes()} nodes")
    print(f"Authority scores: {graph.authority_scores}")
```

#### Legal Concept Tagging

```python
from kite import BAILIIScraper
from kite.utils.legal_concepts import extract_concepts, get_taxonomy

with BAILIIScraper() as scraper:
    cases = scraper.search_cases(query="data protection", limit=5)
    
    for case in cases:
        concepts = extract_concepts(case.text)
        print(f"\n{case.case_name}")
        print(f"  Legal area: {concepts['area_of_law']}")
        print(f"  Concepts: {', '.join(concepts['concepts'][:5])}")
        print(f"  Confidence: {concepts['confidence']:.2f}")
```

#### Data Validation and Quality Metrics

```python
from kite import CanLIIScraper
from kite.utils.validation_pipeline import validate_case, get_quality_metrics

with CanLIIScraper() as scraper:
    case = scraper.get_case_by_id("2023 SCC 15")
    
    # Validate case data
    validation_result = validate_case(case)
    
    if validation_result.is_valid:
        print(f"✅ Case is valid")
        print(f"   Completeness: {validation_result.completeness_score:.2%}")
        print(f"   Quality grade: {validation_result.quality_grade}")
    else:
        print(f"❌ Validation errors:")
        for error in validation_result.errors:
            print(f"   - {error}")
```

#### Jurisdiction Metadata Enrichment

```python
from kite import AustLIIScraper
from kite.utils.jurisdiction_metadata import enrich_metadata

with AustLIIScraper() as scraper:
    case = scraper.search_cases(query="contract law", limit=1)[0]
    
    # Enrich with jurisdiction metadata
    enriched = enrich_metadata(case)
    
    print(f"Court level: {enriched.court_level}")  # e.g., "supreme"
    print(f"Court type: {enriched.court_type}")    # e.g., "appellate"
    print(f"Case type: {enriched.case_type}")      # e.g., "civil"
    print(f"Hierarchy: {enriched.hierarchy_level}") # 1-5
```

#### Ethical Scraping Compliance

```python
from kite.utils.robots_checker import check_robots_txt
from kite.utils.scraping_policy import get_policy
from kite.utils.attribution import generate_attribution

# Check robots.txt compliance
allowed = check_robots_txt("https://www.bailii.org")
print(f"Scraping allowed: {allowed}")

# Get scraping policy
policy = get_policy("bailii")
print(f"Rate limit: {policy.rate_limit}s")
print(f"Respect robots.txt: {policy.respect_robots_txt}")
print(f"Attribution required: {policy.attribution_required}")

# Generate attribution
attribution = generate_attribution("bailii", format="html")
print(attribution)
```

#### Observability - Metrics and Logging

```python
import structlog
from kite import CourtListenerScraper
from kite.utils.metrics import MetricsCollector

logger = structlog.get_logger()
metrics = MetricsCollector()

with CourtListenerScraper() as scraper:
    logger.info("starting_scrape", jurisdiction="US")
    
    cases = scraper.search_cases(
        query="constitutional law",
        limit=10
    )
    
    metrics.record_scrape_success("courtlistener", count=len(cases))
    logger.info("scrape_complete", cases_retrieved=len(cases))

# Metrics available at /metrics endpoint
# Prometheus format
```

#### Batch Processing with Workers

```python
from kite import IndianKanoonScraper
from concurrent.futures import ThreadPoolExecutor
import structlog

logger = structlog.get_logger()

def scrape_case(case_id):
    with IndianKanoonScraper() as scraper:
        try:
            case = scraper.get_case_by_id(case_id)
            logger.info("case_retrieved", case_id=case_id)
            return case
        except Exception as e:
            logger.error("scrape_failed", case_id=case_id, error=str(e))
            return None

case_ids = [f"AIR 2023 SC {i}" for i in range(1000, 1100)]

with ThreadPoolExecutor(max_workers=4) as executor:
    results = list(executor.map(scrape_case, case_ids))
    
successful = [r for r in results if r is not None]
print(f"Retrieved {len(successful)}/{len(case_ids)} cases")
```

### Monitoring

Access Prometheus metrics:

```console
$ curl http://localhost:9090/metrics
```

View Grafana dashboards:

```console
$ open http://localhost:3000
# Default credentials: admin/admin
```
