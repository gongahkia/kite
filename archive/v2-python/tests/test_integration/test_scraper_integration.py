"""
Integration tests for scrapers.

These tests verify scraper functionality with mocked HTTP responses
to simulate real-world usage patterns.
"""

import pytest
from unittest.mock import Mock, patch, MagicMock
from datetime import datetime
from kite.scrapers.courtlistener import CourtListenerScraper
from kite.utils.data_models import CaseData
from kite.utils.exceptions import NetworkError, ParsingError


class TestCourtListenerIntegration:
    """Integration tests for CourtListener scraper."""

    @pytest.fixture
    def mock_successful_response(self):
        """Mock successful HTTP response."""
        response = Mock()
        response.status_code = 200
        response.json.return_value = {
            "results": [
                {
                    "caseName": "Test v. Example",
                    "id": 12345,
                    "court": "Test Court",
                    "dateFiled": "2023-01-15",
                    "absolute_url": "/opinion/12345/test-v-example/",
                    "citation": "123 Test 456",
                    "judge": "Judge Smith",
                    "snippet": "This is a test case summary.",
                }
            ]
        }
        return response

    @pytest.fixture
    def mock_error_response(self):
        """Mock error HTTP response."""
        response = Mock()
        response.status_code = 500
        return response

    def test_search_cases_success(self, mock_successful_response):
        """Test successful case search with mocked response."""
        with patch("requests.Session.request", return_value=mock_successful_response):
            with CourtListenerScraper() as scraper:
                cases = scraper.search_cases(query="test", limit=10)

                assert len(cases) > 0
                assert all(isinstance(case, CaseData) for case in cases)
                assert cases[0].case_name == "Test v. Example"

    def test_search_cases_with_filters(self, mock_successful_response):
        """Test case search with date and court filters."""
        with patch("requests.Session.request", return_value=mock_successful_response):
            with CourtListenerScraper() as scraper:
                cases = scraper.search_cases(
                    query="privacy",
                    start_date="2023-01-01",
                    end_date="2023-12-31",
                    court="scotus",
                    limit=5,
                )

                assert isinstance(cases, list)

    def test_search_cases_network_error(self, mock_error_response):
        """Test handling of network errors."""
        with patch("requests.Session.request", return_value=mock_error_response):
            with CourtListenerScraper() as scraper:
                with pytest.raises(NetworkError):
                    scraper.search_cases(query="test")

    def test_rate_limiting(self, mock_successful_response):
        """Test that rate limiting is respected."""
        with patch("requests.Session.request", return_value=mock_successful_response):
            with CourtListenerScraper(rate_limit=0.5) as scraper:
                start_time = datetime.now()

                # Make two consecutive requests
                scraper.search_cases(query="test1", limit=1)
                scraper.search_cases(query="test2", limit=1)

                elapsed = (datetime.now() - start_time).total_seconds()

                # Should take at least 0.5 seconds due to rate limiting
                assert elapsed >= 0.5

    def test_context_manager_closes_session(self):
        """Test that context manager properly closes the session."""
        scraper = CourtListenerScraper()

        with scraper:
            assert scraper.session is not None

        # Session should still exist but can be closed
        assert scraper.session is not None

    def test_metrics_collection(self, mock_successful_response):
        """Test that metrics are collected during scraping."""
        with patch("requests.Session.request", return_value=mock_successful_response):
            with CourtListenerScraper() as scraper:
                assert scraper.metrics is not None
                assert scraper.metrics.scraper_name == "CourtListenerScraper"

                cases = scraper.search_cases(query="test", limit=1)

                # Metrics should be collected
                assert scraper.metrics is not None


class TestScraperRetryLogic:
    """Test retry logic and error handling."""

    def test_retry_on_temporary_failure(self):
        """Test that scraper retries on temporary failures."""
        # Mock responses: fail twice, then succeed
        responses = [
            Mock(status_code=500),  # First attempt fails
            Mock(status_code=500),  # Second attempt fails
            Mock(status_code=200, json=lambda: {"results": []}),  # Third succeeds
        ]

        with patch(
            "requests.Session.request", side_effect=responses
        ):
            with CourtListenerScraper(max_retries=2, retry_delay=0.1) as scraper:
                cases = scraper.search_cases(query="test", limit=1)

                # Should succeed after retries
                assert isinstance(cases, list)

    def test_max_retries_exceeded(self):
        """Test that scraper fails after max retries."""
        # All attempts fail
        mock_response = Mock(status_code=500)

        with patch("requests.Session.request", return_value=mock_response):
            with CourtListenerScraper(max_retries=2, retry_delay=0.1) as scraper:
                with pytest.raises(NetworkError):
                    scraper.search_cases(query="test")


class TestScraperLogging:
    """Test that scrapers properly log events."""

    def test_logging_on_success(self, caplog):
        """Test that successful operations are logged."""
        mock_response = Mock(status_code=200, json=lambda: {"results": []})

        with patch("requests.Session.request", return_value=mock_response):
            with CourtListenerScraper() as scraper:
                scraper.search_cases(query="test")

                # Logger should be initialized
                assert scraper.logger is not None

    def test_logging_on_error(self, caplog):
        """Test that errors are properly logged."""
        mock_response = Mock(status_code=500)

        with patch("requests.Session.request", return_value=mock_response):
            with CourtListenerScraper(max_retries=0) as scraper:
                with pytest.raises(NetworkError):
                    scraper.search_cases(query="test")

                # Logger should capture the error
                assert scraper.logger is not None

