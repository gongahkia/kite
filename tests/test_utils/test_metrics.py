"""
Tests for metrics collection.
"""

import pytest
import time
from kite.utils.metrics import (
    MetricsCollector,
    increment_request_counter,
    record_request_duration,
    increment_error_counter,
)


class TestMetricsCollector:
    """Test metrics collector functionality."""

    def test_init(self):
        """Test metrics collector initialization."""
        collector = MetricsCollector("TestScraper", "US")

        assert collector.scraper_name == "TestScraper"
        assert collector.jurisdiction == "US"

    def test_track_request_success(self):
        """Test tracking successful HTTP request."""
        collector = MetricsCollector("TestScraper")

        with collector.track_request("GET") as req:
            req["status"] = 200

        # Should not raise

    def test_track_request_with_error(self):
        """Test tracking request with error."""
        collector = MetricsCollector("TestScraper")

        with pytest.raises(ValueError):
            with collector.track_request("GET") as req:
                raise ValueError("Test error")

    def test_track_scraping(self):
        """Test tracking scraping operation."""
        collector = MetricsCollector("TestScraper", "US")

        with collector.track_scraping() as scrape:
            scrape["cases_count"] = 5

        # Should not raise

    def test_track_parsing(self):
        """Test tracking parsing operation."""
        collector = MetricsCollector("TestScraper")

        with collector.track_parsing() as parse:
            # Simulate parsing
            pass

        # Should not raise

    def test_track_parsing_with_error(self):
        """Test tracking parsing with error."""
        collector = MetricsCollector("TestScraper")

        with pytest.raises(RuntimeError):
            with collector.track_parsing() as parse:
                raise RuntimeError("Parsing failed")

    def test_record_rate_limit(self):
        """Test recording rate limit event."""
        collector = MetricsCollector("TestScraper")

        # Should not raise
        collector.record_rate_limit(1.5)


class TestMetricsUtilities:
    """Test metrics utility functions."""

    def test_increment_request_counter(self):
        """Test incrementing request counter."""
        # Should not raise
        increment_request_counter("TestScraper", "GET", 200)

    def test_record_request_duration(self):
        """Test recording request duration."""
        # Should not raise
        record_request_duration("TestScraper", "GET", 0.5)

    def test_increment_error_counter(self):
        """Test incrementing error counter."""
        # Should not raise
        increment_error_counter("TestScraper", "NetworkError")

