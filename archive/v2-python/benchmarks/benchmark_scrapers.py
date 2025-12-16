"""Basic performance benchmarks for scrapers."""
import time
from kite import CourtListenerScraper

def benchmark_search():
    """Benchmark search performance."""
    scraper = CourtListenerScraper()
    start = time.time()
    results = scraper.search_cases("test", limit=10)
    duration = time.time() - start
    print(f"Search took {duration:.2f}s, found {len(results)} cases")

if __name__ == "__main__":
    benchmark_search()
