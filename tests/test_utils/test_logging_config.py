"""
Tests for structured logging configuration.
"""

import pytest
import structlog
from kite.utils.logging_config import (
    configure_structlog,
    get_logger,
    log_context,
    bind_context,
    clear_context,
)


class TestLoggingConfiguration:
    """Test structured logging configuration."""

    def test_configure_structlog_json(self):
        """Test JSON format configuration."""
        configure_structlog(log_level="INFO", log_format="json")

        # Should not raise
        logger = structlog.get_logger("test")
        assert logger is not None

    def test_configure_structlog_console(self):
        """Test console format configuration."""
        configure_structlog(log_level="DEBUG", log_format="console")

        # Should not raise
        logger = structlog.get_logger("test")
        assert logger is not None

    def test_configure_structlog_text(self):
        """Test text format configuration."""
        configure_structlog(log_level="WARNING", log_format="text")

        # Should not raise
        logger = structlog.get_logger("test")
        assert logger is not None

    def test_get_logger(self):
        """Test getting a configured logger."""
        logger = get_logger("test")

        assert logger is not None
        assert hasattr(logger, "info")
        assert hasattr(logger, "warning")
        assert hasattr(logger, "error")

    def test_logger_with_context(self):
        """Test logger with context."""
        logger = get_logger("test")

        # Should not raise
        logger.info("test_event", key="value", number=42)

    def test_log_context_manager(self):
        """Test log context manager."""
        with log_context(operation="test", user_id=123):
            logger = get_logger("test")
            # Context should be available
            logger.info("test_event")

        # Context should be cleared after exiting
        clear_context()

    def test_bind_and_unbind_context(self):
        """Test binding and unbinding context."""
        bind_context(request_id="abc123", user="test")

        logger = get_logger("test")
        logger.info("test_event")

        clear_context()

    def test_clear_context(self):
        """Test clearing context."""
        bind_context(key="value")
        clear_context()

        # Context should be empty now
        logger = get_logger("test")
        logger.info("test_event")

