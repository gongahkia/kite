# Version 3 (Nim)

Version 3 is the Nim rewrite providing a static binary, fast startup, and parity with prior versions.

- Language: Nim (latest stable)
- Sources: currently under repository `src/` (will be the canonical v3 location)
- Binary: `kite`
- Metrics: Prometheus at /metrics; health at /health

This folder documents v3; sources are maintained at the repository root `src/` to avoid breaking imports and build paths.

## Usage

### Setup

#### Build from Source

```console
$ cd archive/v3-nim
$ nimble install
$ nimble build
```

This creates a `kite` binary in the project directory.

#### Install System-Wide

```console
$ nimble install
```

### CLI Usage

#### List Available Scrapers

```console
$ ./kite list-scrapers
```

#### Search for Cases

```console
$ ./kite search courtlistener "privacy rights" --limit 10
$ ./kite search bailii "contract law" -l 5 -f json
$ ./kite search canlii "constitutional rights" --start-date 2023-01-01 --court "Supreme Court"
```

With output options:

```console
$ ./kite search austlii "tort law" --format json --output results.json
$ ./kite search worldlii "human rights" -f csv -o cases.csv
```

#### Get Specific Case by ID

```console
$ ./kite get-case canlii "2023 SCC 15"
$ ./kite get-case bailii "EWCA/Civ/2023/1234" --format json
$ ./kite get-case courtlistener "12345" -o case.json
```

#### Version Information

```console
$ ./kite --version
```

### Library Usage (Nim)

You can import and use Kite v3 as a Nim library:

#### Search Cases Across Jurisdictions

```nim
import kite/base_scraper
import kite/scrapers/courtlistener_scraper
import kite/data_models

let scraper = newCourtListenerScraper()
let query = SearchQuery(
  query: "constitutional law",
  limit: 10,
  startDate: "2023-01-01"
)

let cases = scraper.searchCases(query)
for case in cases:
  echo "Case: ", case.caseName
  echo "Date: ", case.date
  echo "Court: ", case.court
  echo "---"
```

#### Retrieve Specific Case

```nim
import kite/scrapers/canlii_scraper

let scraper = newCanLIIScraper()
let case = scraper.getCaseById("2023 SCC 15")

if case.isSome:
  let c = case.get()
  echo "Case Name: ", c.caseName
  echo "Judges: ", c.judges.join(", ")
  echo "URL: ", c.url
else:
  echo "Case not found"
```

#### Multi-jurisdiction Research

```nim
import kite/scrapers/[courtlistener_scraper, bailii_scraper, austlii_scraper]
import sequtils

type JurisdictionScraper = tuple[name: string, scraper: BaseScraper]

let scrapers: seq[JurisdictionScraper] = @[
  ("US", newCourtListenerScraper()),
  ("UK", newBAILIIScraper()),
  ("AU", newAustLIIScraper())
]

let query = "data protection"
var allCases: seq[CaseData] = @[]

for (jurisdiction, scraper) in scrapers:
  let cases = scraper.searchCases(SearchQuery(query: query, limit: 5))
  for case in cases:
    var c = case
    c.metadata["jurisdiction"] = jurisdiction
    allCases.add(c)

echo "Total cases retrieved: ", allCases.len
```

#### Error Handling

```nim
import kite/scrapers/indian_kanoon_scraper
import kite/exceptions

let scraper = newIndianKanoonScraper()
let caseIds = @["AIR 2023 SC 1234", "AIR 2023 SC 5678"]

for caseId in caseIds:
  try:
    let case = scraper.getCaseById(caseId)
    if case.isSome:
      echo "Retrieved: ", case.get().caseName
  except ScrapingError as e:
    echo "Failed to retrieve ", caseId, ": ", e.msg
  except NetworkError as e:
    echo "Network error for ", caseId, ": ", e.msg
```

#### Configuration and Logging

```nim
import kite/config
import kite/logging_config
import chronicles

# Configure logging
setupLogging(level = LogLevel.INFO)

# Load configuration
let config = loadConfig("config.yaml")

info "Starting scrape", jurisdiction = "US"
let scraper = newCourtListenerScraper(config.courtlistener)
let cases = scraper.searchCases(SearchQuery(query: "tort law", limit: 5))
info "Scrape complete", casesCount = cases.len
```

#### Custom Rate Limiting

```nim
import kite/base_scraper
import kite/scrapers/bailii_scraper

let scraper = newBAILIIScraper(
  rateLimit = 2.0,      # 2 seconds between requests
  timeout = 30,         # 30 second timeout
  maxRetries = 5        # retry up to 5 times
)

let cases = scraper.searchCases(SearchQuery(query: "negligence", limit: 20))
```

### Performance Comparison

Version 3 provides significant performance improvements:

```console
# Binary size
$ ls -lh kite
-rwxr-xr-x  1 user  staff   2.1M  kite

# Startup time
$ time ./kite --version
Kite 2.0.0 (Nim)

real    0m0.003s
user    0m0.001s
sys     0m0.001s
```

### Metrics Endpoint

Kite v3 exposes Prometheus metrics:

```console
$ curl http://localhost:9090/metrics
# TYPE kite_scrapes_total counter
kite_scrapes_total{jurisdiction="courtlistener",status="success"} 152
kite_scrapes_total{jurisdiction="bailii",status="success"} 89
# TYPE kite_scrape_duration_seconds histogram
kite_scrape_duration_seconds_bucket{jurisdiction="canlii",le="1.0"} 45
```

### Docker Deployment

```console
$ cd archive/v3-nim
$ docker build -t kite:v3 .
$ docker run -p 9090:9090 kite:v3
```

### Cross-Compilation

Compile for different platforms:

```console
# Linux
$ nim c -d:release --os:linux --cpu:amd64 src/kite.nim

# Windows
$ nim c -d:release --os:windows --cpu:amd64 src/kite.nim

# macOS
$ nim c -d:release --os:macosx --cpu:amd64 src/kite.nim
```
