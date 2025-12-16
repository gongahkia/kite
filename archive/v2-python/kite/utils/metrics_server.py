"""
Metrics server for exposing Prometheus metrics.

Runs a simple HTTP server to expose metrics endpoint.
"""

import sys
import signal
from http.server import HTTPServer, BaseHTTPRequestHandler
from prometheus_client import generate_latest, CONTENT_TYPE_LATEST
from typing import Any

from .config import get_config
from .logging_config import get_logger

logger = get_logger(__name__)


class MetricsHandler(BaseHTTPRequestHandler):
    """HTTP handler for metrics endpoint."""

    def do_GET(self) -> None:
        """Handle GET requests."""
        config = get_config()

        if self.path == config.metrics_path:
            # Serve metrics
            self.send_response(200)
            self.send_header("Content-Type", CONTENT_TYPE_LATEST)
            self.end_headers()
            self.wfile.write(generate_latest())

        elif self.path == "/health":
            # Health check endpoint
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.end_headers()
            response = b'{"status": "healthy", "service": "kite"}'
            self.wfile.write(response)

        elif self.path == "/":
            # Root path - redirect to metrics
            self.send_response(302)
            self.send_header("Location", config.metrics_path)
            self.end_headers()

        else:
            # Not found
            self.send_response(404)
            self.end_headers()
            self.wfile.write(b"Not Found")

    def log_message(self, format: str, *args: Any) -> None:
        """Override to use structured logging."""
        logger.info("metrics_request", path=self.path, method=self.command)


def run_server() -> None:
    """
    Run the metrics server.

    Starts HTTP server on configured port and exposes metrics endpoint.
    """
    config = get_config()
    port = config.metrics_port

    server = HTTPServer(("0.0.0.0", port), MetricsHandler)

    # Handle shutdown gracefully
    def signal_handler(signum: int, frame: Any) -> None:
        logger.info("shutdown_signal_received", signal=signum)
        server.shutdown()
        sys.exit(0)

    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)

    logger.info(
        "metrics_server_started",
        port=port,
        metrics_path=config.metrics_path,
        health_path="/health",
    )

    try:
        server.serve_forever()
    except KeyboardInterrupt:
        logger.info("metrics_server_stopped")
    finally:
        server.server_close()


if __name__ == "__main__":
    run_server()

