"""
Structured logging configuration using structlog.

Provides JSON-formatted, context-rich logging for better observability.
"""

import sys
import logging
from typing import Optional
import structlog
from structlog.types import EventDict, WrappedLogger
from pathlib import Path

from .config import get_config


def add_app_context(
    logger: WrappedLogger, method_name: str, event_dict: EventDict
) -> EventDict:
    """
    Add application context to log events.

    Args:
        logger: The wrapped logger
        method_name: The name of the method being called
        event_dict: The event dictionary

    Returns:
        EventDict: Enhanced event dictionary
    """
    event_dict["app"] = "kite"
    event_dict["version"] = "1.0.0"
    return event_dict


def configure_structlog(log_level: str = "INFO", log_format: str = "json") -> None:
    """
    Configure structlog for the application.

    Args:
        log_level: Logging level (DEBUG, INFO, WARNING, ERROR, CRITICAL)
        log_format: Output format (json, console, text)
    """
    # Set up standard library logging
    logging.basicConfig(
        format="%(message)s",
        stream=sys.stdout,
        level=getattr(logging, log_level.upper()),
    )

    # Configure processors based on format
    if log_format == "json":
        processors = [
            structlog.contextvars.merge_contextvars,
            structlog.stdlib.add_logger_name,
            structlog.stdlib.add_log_level,
            structlog.stdlib.PositionalArgumentsFormatter(),
            add_app_context,
            structlog.processors.TimeStamper(fmt="iso"),
            structlog.processors.StackInfoRenderer(),
            structlog.processors.format_exc_info,
            structlog.processors.UnicodeDecoder(),
            structlog.processors.JSONRenderer(),
        ]
    elif log_format == "console":
        processors = [
            structlog.contextvars.merge_contextvars,
            structlog.stdlib.add_logger_name,
            structlog.stdlib.add_log_level,
            structlog.stdlib.PositionalArgumentsFormatter(),
            add_app_context,
            structlog.processors.TimeStamper(fmt="%Y-%m-%d %H:%M:%S"),
            structlog.processors.StackInfoRenderer(),
            structlog.processors.format_exc_info,
            structlog.dev.ConsoleRenderer(),
        ]
    else:  # text format
        processors = [
            structlog.contextvars.merge_contextvars,
            structlog.stdlib.add_logger_name,
            structlog.stdlib.add_log_level,
            structlog.stdlib.PositionalArgumentsFormatter(),
            add_app_context,
            structlog.processors.TimeStamper(fmt="iso"),
            structlog.processors.StackInfoRenderer(),
            structlog.processors.format_exc_info,
            structlog.processors.UnicodeDecoder(),
            structlog.processors.KeyValueRenderer(),
        ]

    structlog.configure(
        processors=processors,
        wrapper_class=structlog.stdlib.BoundLogger,
        context_class=dict,
        logger_factory=structlog.stdlib.LoggerFactory(),
        cache_logger_on_first_use=True,
    )


def get_logger(name: Optional[str] = None) -> structlog.BoundLogger:
    """
    Get a configured structlog logger.

    Args:
        name: Logger name (typically __name__)

    Returns:
        structlog.BoundLogger: Configured logger instance
    """
    # Ensure structlog is configured
    try:
        config = get_config()
        configure_structlog(config.log_level, config.log_format)
    except Exception:
        # Fallback to default configuration
        configure_structlog()

    return structlog.get_logger(name)


def setup_file_logging(log_file_path: str) -> None:
    """
    Set up file-based logging in addition to console logging.

    Args:
        log_file_path: Path to log file
    """
    # Create log directory if it doesn't exist
    log_path = Path(log_file_path)
    log_path.parent.mkdir(parents=True, exist_ok=True)

    # Add file handler to root logger
    file_handler = logging.FileHandler(log_file_path)
    file_handler.setLevel(logging.DEBUG)

    # Use JSON format for file logs
    formatter = logging.Formatter("%(message)s")
    file_handler.setFormatter(formatter)

    root_logger = logging.getLogger()
    root_logger.addHandler(file_handler)


# Context managers for log context
class log_context:
    """
    Context manager for adding context to logs.

    Example:
        >>> with log_context(scraper="CourtListener", operation="search"):
        ...     logger.info("performing_search", query="privacy")
    """

    def __init__(self, **kwargs):
        """
        Initialize log context.

        Args:
            **kwargs: Context key-value pairs to add to logs
        """
        self.context = kwargs

    def __enter__(self):
        """Enter context."""
        structlog.contextvars.clear_contextvars()
        structlog.contextvars.bind_contextvars(**self.context)
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """Exit context."""
        structlog.contextvars.clear_contextvars()


def bind_context(**kwargs) -> None:
    """
    Bind context variables that will be included in all subsequent log messages.

    Args:
        **kwargs: Context key-value pairs
    """
    structlog.contextvars.bind_contextvars(**kwargs)


def clear_context() -> None:
    """Clear all bound context variables."""
    structlog.contextvars.clear_contextvars()


def unbind_context(*keys: str) -> None:
    """
    Unbind specific context variables.

    Args:
        *keys: Context keys to unbind
    """
    structlog.contextvars.unbind_contextvars(*keys)

