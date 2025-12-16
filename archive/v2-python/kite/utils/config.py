"""
Configuration management for Kite library.
"""

import os
from typing import Optional
from dataclasses import dataclass


@dataclass
class Config:
    """
    Configuration settings for Kite.

    Loads configuration from environment variables with sensible defaults.
    """

    # Application settings
    python_env: str = "development"
    log_level: str = "INFO"

    # Scraper settings
    rate_limit: float = 1.0
    request_timeout: int = 30
    max_retries: int = 3
    user_agent: Optional[str] = None

    # Metrics & monitoring
    metrics_port: int = 8000
    metrics_path: str = "/metrics"
    enable_metrics: bool = True
    enable_tracing: bool = True

    # Logging configuration
    log_format: str = "json"
    log_file_path: str = "./logs/app.log"
    log_rotation: str = "10MB"
    log_retention: int = 30

    @classmethod
    def from_env(cls) -> "Config":
        """
        Load configuration from environment variables.

        Returns:
            Config: Configuration instance with values from environment
        """
        return cls(
            python_env=os.getenv("PYTHON_ENV", "development"),
            log_level=os.getenv("LOG_LEVEL", "INFO"),
            rate_limit=float(os.getenv("RATE_LIMIT", "1.0")),
            request_timeout=int(os.getenv("REQUEST_TIMEOUT", "30")),
            max_retries=int(os.getenv("MAX_RETRIES", "3")),
            user_agent=os.getenv("USER_AGENT"),
            metrics_port=int(os.getenv("METRICS_PORT", "8000")),
            metrics_path=os.getenv("METRICS_PATH", "/metrics"),
            enable_metrics=os.getenv("ENABLE_METRICS", "true").lower() == "true",
            enable_tracing=os.getenv("ENABLE_TRACING", "true").lower() == "true",
            log_format=os.getenv("LOG_FORMAT", "json"),
            log_file_path=os.getenv("LOG_FILE_PATH", "./logs/app.log"),
            log_rotation=os.getenv("LOG_ROTATION", "10MB"),
            log_retention=int(os.getenv("LOG_RETENTION", "30")),
        )

    def is_production(self) -> bool:
        """Check if running in production environment."""
        return self.python_env.lower() == "production"

    def is_development(self) -> bool:
        """Check if running in development environment."""
        return self.python_env.lower() == "development"

    def validate(self) -> None:
        """
        Validate configuration values.

        Raises:
            ValueError: If configuration is invalid
        """
        if self.rate_limit < 0:
            raise ValueError("rate_limit must be non-negative")

        if self.request_timeout <= 0:
            raise ValueError("request_timeout must be positive")

        if self.max_retries < 0:
            raise ValueError("max_retries must be non-negative")

        if self.metrics_port < 1 or self.metrics_port > 65535:
            raise ValueError("metrics_port must be between 1 and 65535")

        valid_log_levels = ["DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"]
        if self.log_level.upper() not in valid_log_levels:
            raise ValueError(f"log_level must be one of {valid_log_levels}")

        valid_log_formats = ["json", "text", "console"]
        if self.log_format.lower() not in valid_log_formats:
            raise ValueError(f"log_format must be one of {valid_log_formats}")


# Global configuration instance
_config: Optional[Config] = None


def get_config() -> Config:
    """
    Get the global configuration instance.

    Returns:
        Config: Global configuration instance
    """
    global _config
    if _config is None:
        _config = Config.from_env()
        _config.validate()
    return _config


def reset_config() -> None:
    """Reset the global configuration instance (useful for testing)."""
    global _config
    _config = None

