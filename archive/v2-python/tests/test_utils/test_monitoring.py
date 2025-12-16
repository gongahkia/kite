"""
Tests for monitoring and health checks.
"""

import pytest
from kite.utils.monitoring import (
    HealthChecker,
    HealthStatus,
    get_health_checker,
    perform_health_check,
)


class TestHealthChecker:
    """Test health checker functionality."""

    def test_init(self):
        """Test health checker initialization."""
        checker = HealthChecker()

        assert checker.start_time is not None
        assert checker.last_check is None

    def test_get_uptime(self):
        """Test uptime calculation."""
        checker = HealthChecker()
        uptime = checker.get_uptime()

        assert uptime >= 0
        assert isinstance(uptime, float)

    def test_check_system_resources(self):
        """Test system resource checks."""
        checker = HealthChecker()
        resources = checker.check_system_resources()

        assert "status" in resources
        assert "cpu_percent" in resources
        assert "memory_percent" in resources

    def test_check_dependencies(self):
        """Test dependency checks."""
        checker = HealthChecker()
        deps = checker.check_dependencies()

        assert "status" in deps
        assert "dependencies" in deps

        # Check that required dependencies are present
        dependencies = deps["dependencies"]
        assert "requests" in dependencies
        assert "beautifulsoup4" in dependencies
        assert "lxml" in dependencies
        assert "structlog" in dependencies

    def test_perform_health_check(self):
        """Test comprehensive health check."""
        checker = HealthChecker()
        status = checker.perform_health_check()

        assert isinstance(status, HealthStatus)
        assert status.status in ["healthy", "degraded", "unhealthy"]
        assert status.timestamp is not None
        assert status.uptime_seconds >= 0
        assert "system" in status.checks
        assert "dependencies" in status.checks
        assert "uptime" in status.checks

    def test_is_healthy(self):
        """Test health status check."""
        checker = HealthChecker()
        is_healthy = checker.is_healthy()

        assert isinstance(is_healthy, bool)

    def test_format_uptime(self):
        """Test uptime formatting."""
        checker = HealthChecker()

        # Test various durations
        assert "0s" in checker._format_uptime(0)
        assert "5s" in checker._format_uptime(5)
        assert "1m" in checker._format_uptime(60)
        assert "1h" in checker._format_uptime(3600)
        assert "1d" in checker._format_uptime(86400)


class TestHealthStatus:
    """Test health status data class."""

    def test_to_dict(self):
        """Test converting health status to dictionary."""
        status = HealthStatus(
            status="healthy",
            timestamp="2024-01-01T00:00:00",
            uptime_seconds=100.0,
            checks={"test": {"status": "healthy"}},
        )

        status_dict = status.to_dict()

        assert isinstance(status_dict, dict)
        assert status_dict["status"] == "healthy"
        assert status_dict["uptime_seconds"] == 100.0


class TestMonitoringUtilities:
    """Test monitoring utility functions."""

    def test_get_health_checker_singleton(self):
        """Test that get_health_checker returns a singleton."""
        checker1 = get_health_checker()
        checker2 = get_health_checker()

        assert checker1 is checker2

    def test_perform_health_check_function(self):
        """Test convenience health check function."""
        result = perform_health_check()

        assert isinstance(result, dict)
        assert "status" in result
        assert "timestamp" in result
        assert "checks" in result

