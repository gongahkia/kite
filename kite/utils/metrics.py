"""
Metrics collection for Kite library.

Provides Prometheus-compatible metrics for monitoring scraper performance
and health.
"""

from typing import Dict, Optional, Any
from prometheus_client import Counter, Histogram, Gauge, Info
import time
from functools import wraps
from contextlib import contextmanager


# ===== Request Metrics =====
request_counter = Counter(
    "kite_requests_total",
    "Total number of HTTP requests made",
    ["scraper", "method", "status"],
)

request_duration = Histogram(
    "kite_request_duration_seconds",
    "HTTP request duration in seconds",
    ["scraper", "method"],
    buckets=(0.1, 0.5, 1.0, 2.0, 5.0, 10.0, 30.0, 60.0),
)

request_errors = Counter(
    "kite_request_errors_total",
    "Total number of request errors",
    ["scraper", "error_type"],
)

# ===== Scraper Metrics =====
cases_scraped = Counter(
    "kite_cases_scraped_total",
    "Total number of cases successfully scraped",
    ["scraper", "jurisdiction"],
)

scraping_duration = Histogram(
    "kite_scraping_duration_seconds",
    "Time taken to scrape cases",
    ["scraper"],
    buckets=(1.0, 5.0, 10.0, 30.0, 60.0, 120.0, 300.0),
)

active_scrapers = Gauge(
    "kite_active_scrapers",
    "Number of currently active scrapers",
    ["scraper"],
)

# ===== Rate Limiting Metrics =====
rate_limit_hits = Counter(
    "kite_rate_limit_hits_total",
    "Number of times rate limiting was triggered",
    ["scraper"],
)

rate_limit_wait_time = Histogram(
    "kite_rate_limit_wait_seconds",
    "Time spent waiting due to rate limiting",
    ["scraper"],
    buckets=(0.1, 0.5, 1.0, 2.0, 5.0, 10.0),
)

# ===== Parsing Metrics =====
parsing_errors = Counter(
    "kite_parsing_errors_total",
    "Total number of parsing errors",
    ["scraper", "error_type"],
)

parsing_duration = Histogram(
    "kite_parsing_duration_seconds",
    "Time taken to parse case data",
    ["scraper"],
    buckets=(0.01, 0.05, 0.1, 0.5, 1.0, 5.0),
)

# ===== Application Info =====
app_info = Info("kite_app", "Application information")
app_info.info(
    {
        "version": "1.0.0",
        "name": "Kite",
        "description": "Legal case law scraper library",
    }
)


class MetricsCollector:
    """
    Metrics collector for tracking scraper operations.

    Provides context managers and decorators for automatic metrics collection.
    """

    def __init__(self, scraper_name: str, jurisdiction: Optional[str] = None):
        """
        Initialize metrics collector.

        Args:
            scraper_name: Name of the scraper
            jurisdiction: Jurisdiction being scraped
        """
        self.scraper_name = scraper_name
        self.jurisdiction = jurisdiction or "unknown"

    @contextmanager
    def track_request(
        self, method: str = "GET", url: Optional[str] = None
    ) -> Any:
        """
        Context manager to track HTTP request metrics.

        Args:
            method: HTTP method
            url: Request URL (optional)

        Yields:
            Dict with request metadata

        Example:
            >>> collector = MetricsCollector("CourtListener")
            >>> with collector.track_request("GET") as req:
            ...     response = make_request()
            ...     req["status"] = response.status_code
        """
        start_time = time.time()
        request_data = {"status": "unknown", "error": None}

        try:
            yield request_data
            status = request_data.get("status", "unknown")
            request_counter.labels(
                scraper=self.scraper_name, method=method, status=status
            ).inc()

        except Exception as e:
            request_data["error"] = type(e).__name__
            request_errors.labels(
                scraper=self.scraper_name, error_type=type(e).__name__
            ).inc()
            raise

        finally:
            duration = time.time() - start_time
            request_duration.labels(scraper=self.scraper_name, method=method).observe(
                duration
            )

    @contextmanager
    def track_scraping(self) -> Any:
        """
        Context manager to track scraping operation metrics.

        Yields:
            Dict with scraping metadata

        Example:
            >>> collector = MetricsCollector("CourtListener", "US")
            >>> with collector.track_scraping() as scrape:
            ...     cases = scrape_cases()
            ...     scrape["cases_count"] = len(cases)
        """
        start_time = time.time()
        scrape_data = {"cases_count": 0}

        # Increment active scrapers
        active_scrapers.labels(scraper=self.scraper_name).inc()

        try:
            yield scrape_data

            # Record successful scraping
            count = scrape_data.get("cases_count", 0)
            if count > 0:
                cases_scraped.labels(
                    scraper=self.scraper_name, jurisdiction=self.jurisdiction
                ).inc(count)

        finally:
            # Decrement active scrapers
            active_scrapers.labels(scraper=self.scraper_name).dec()

            # Record duration
            duration = time.time() - start_time
            scraping_duration.labels(scraper=self.scraper_name).observe(duration)

    @contextmanager
    def track_parsing(self) -> Any:
        """
        Context manager to track parsing operation metrics.

        Yields:
            Dict with parsing metadata

        Example:
            >>> collector = MetricsCollector("CourtListener")
            >>> with collector.track_parsing() as parse:
            ...     case_data = parse_html(html)
        """
        start_time = time.time()
        parse_data = {"error": None}

        try:
            yield parse_data

        except Exception as e:
            parse_data["error"] = type(e).__name__
            parsing_errors.labels(
                scraper=self.scraper_name, error_type=type(e).__name__
            ).inc()
            raise

        finally:
            duration = time.time() - start_time
            parsing_duration.labels(scraper=self.scraper_name).observe(duration)

    def record_rate_limit(self, wait_time: float) -> None:
        """
        Record rate limiting event.

        Args:
            wait_time: Time spent waiting (seconds)
        """
        rate_limit_hits.labels(scraper=self.scraper_name).inc()
        rate_limit_wait_time.labels(scraper=self.scraper_name).observe(wait_time)


def track_function_duration(scraper_name: str, operation: str = "operation"):
    """
    Decorator to track function execution duration.

    Args:
        scraper_name: Name of the scraper
        operation: Name of the operation being tracked

    Example:
        >>> @track_function_duration("CourtListener", "search")
        ... def search_cases(query):
        ...     return perform_search(query)
    """

    def decorator(func):
        @wraps(func)
        def wrapper(*args, **kwargs):
            start_time = time.time()
            try:
                result = func(*args, **kwargs)
                return result
            finally:
                duration = time.time() - start_time
                scraping_duration.labels(scraper=scraper_name).observe(duration)

        return wrapper

    return decorator


# Utility functions for common metrics operations
def increment_request_counter(
    scraper_name: str, method: str = "GET", status: int = 200
) -> None:
    """
    Increment request counter.

    Args:
        scraper_name: Name of the scraper
        method: HTTP method
        status: HTTP status code
    """
    request_counter.labels(
        scraper=scraper_name, method=method, status=str(status)
    ).inc()


def record_request_duration(
    scraper_name: str, method: str, duration: float
) -> None:
    """
    Record request duration.

    Args:
        scraper_name: Name of the scraper
        method: HTTP method
        duration: Duration in seconds
    """
    request_duration.labels(scraper=scraper_name, method=method).observe(duration)


def increment_error_counter(scraper_name: str, error_type: str) -> None:
    """
    Increment error counter.

    Args:
        scraper_name: Name of the scraper
        error_type: Type of error
    """
    request_errors.labels(scraper=scraper_name, error_type=error_type).inc()

