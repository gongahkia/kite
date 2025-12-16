"""
Monitoring and health check utilities for Kite library.
"""

import time
import psutil
from typing import Dict, Any, Optional
from datetime import datetime
from dataclasses import dataclass, asdict


@dataclass
class HealthStatus:
    """
    Health check status information.
    """

    status: str  # "healthy", "degraded", "unhealthy"
    timestamp: str
    uptime_seconds: float
    checks: Dict[str, Any]

    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary."""
        return asdict(self)


class HealthChecker:
    """
    Health checker for monitoring application health.

    Performs various health checks and aggregates results.
    """

    def __init__(self):
        """Initialize health checker."""
        self.start_time = time.time()
        self.last_check: Optional[HealthStatus] = None

    def get_uptime(self) -> float:
        """
        Get application uptime in seconds.

        Returns:
            float: Uptime in seconds
        """
        return time.time() - self.start_time

    def check_system_resources(self) -> Dict[str, Any]:
        """
        Check system resource usage.

        Returns:
            Dict with resource usage information
        """
        try:
            cpu_percent = psutil.cpu_percent(interval=0.1)
            memory = psutil.virtual_memory()
            disk = psutil.disk_usage("/")

            return {
                "status": "healthy",
                "cpu_percent": cpu_percent,
                "memory_percent": memory.percent,
                "memory_available_mb": memory.available / (1024 * 1024),
                "disk_percent": disk.percent,
                "disk_available_gb": disk.free / (1024 * 1024 * 1024),
            }
        except Exception as e:
            return {"status": "unhealthy", "error": str(e)}

    def check_dependencies(self) -> Dict[str, Any]:
        """
        Check if required dependencies are available.

        Returns:
            Dict with dependency status
        """
        checks = {}

        # Check requests library
        try:
            import requests

            checks["requests"] = {"status": "available", "version": requests.__version__}
        except ImportError as e:
            checks["requests"] = {"status": "missing", "error": str(e)}

        # Check BeautifulSoup
        try:
            import bs4

            checks["beautifulsoup4"] = {"status": "available", "version": bs4.__version__}
        except ImportError as e:
            checks["beautifulsoup4"] = {"status": "missing", "error": str(e)}

        # Check lxml
        try:
            import lxml

            checks["lxml"] = {"status": "available", "version": lxml.__version__}
        except ImportError as e:
            checks["lxml"] = {"status": "missing", "error": str(e)}

        # Check structlog
        try:
            import structlog

            checks["structlog"] = {
                "status": "available",
                "version": structlog.__version__,
            }
        except ImportError as e:
            checks["structlog"] = {"status": "missing", "error": str(e)}

        # Determine overall status
        all_available = all(
            check.get("status") == "available" for check in checks.values()
        )
        status = "healthy" if all_available else "degraded"

        return {"status": status, "dependencies": checks}

    def perform_health_check(self) -> HealthStatus:
        """
        Perform comprehensive health check.

        Returns:
            HealthStatus: Overall health status
        """
        checks = {
            "system": self.check_system_resources(),
            "dependencies": self.check_dependencies(),
            "uptime": {
                "status": "healthy",
                "uptime_seconds": self.get_uptime(),
                "uptime_human": self._format_uptime(self.get_uptime()),
            },
        }

        # Determine overall status
        statuses = [check.get("status", "unknown") for check in checks.values()]

        if all(s == "healthy" for s in statuses):
            overall_status = "healthy"
        elif any(s == "unhealthy" for s in statuses):
            overall_status = "unhealthy"
        else:
            overall_status = "degraded"

        health_status = HealthStatus(
            status=overall_status,
            timestamp=datetime.utcnow().isoformat(),
            uptime_seconds=self.get_uptime(),
            checks=checks,
        )

        self.last_check = health_status
        return health_status

    def _format_uptime(self, seconds: float) -> str:
        """
        Format uptime in human-readable format.

        Args:
            seconds: Uptime in seconds

        Returns:
            str: Formatted uptime string
        """
        days = int(seconds // 86400)
        hours = int((seconds % 86400) // 3600)
        minutes = int((seconds % 3600) // 60)
        secs = int(seconds % 60)

        parts = []
        if days > 0:
            parts.append(f"{days}d")
        if hours > 0:
            parts.append(f"{hours}h")
        if minutes > 0:
            parts.append(f"{minutes}m")
        parts.append(f"{secs}s")

        return " ".join(parts)

    def is_healthy(self) -> bool:
        """
        Quick health check.

        Returns:
            bool: True if system is healthy
        """
        if self.last_check is None:
            status = self.perform_health_check()
        else:
            status = self.last_check

        return status.status == "healthy"


# Global health checker instance
_health_checker: Optional[HealthChecker] = None


def get_health_checker() -> HealthChecker:
    """
    Get the global health checker instance.

    Returns:
        HealthChecker: Global health checker
    """
    global _health_checker
    if _health_checker is None:
        _health_checker = HealthChecker()
    return _health_checker


def perform_health_check() -> Dict[str, Any]:
    """
    Convenience function to perform health check.

    Returns:
        Dict: Health check results
    """
    checker = get_health_checker()
    status = checker.perform_health_check()
    return status.to_dict()

