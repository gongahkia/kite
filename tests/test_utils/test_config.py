"""
Tests for configuration management.
"""

import pytest
import os
from kite.utils.config import Config, get_config, reset_config


class TestConfig:
    """Test configuration management."""

    def setup_method(self):
        """Reset configuration before each test."""
        reset_config()

    def test_default_config(self):
        """Test default configuration values."""
        config = Config()

        assert config.python_env == "development"
        assert config.log_level == "INFO"
        assert config.rate_limit == 1.0
        assert config.request_timeout == 30
        assert config.max_retries == 3
        assert config.metrics_port == 8000
        assert config.enable_metrics is True

    def test_config_from_env(self, monkeypatch):
        """Test loading configuration from environment variables."""
        monkeypatch.setenv("PYTHON_ENV", "production")
        monkeypatch.setenv("LOG_LEVEL", "DEBUG")
        monkeypatch.setenv("RATE_LIMIT", "2.5")
        monkeypatch.setenv("REQUEST_TIMEOUT", "60")
        monkeypatch.setenv("METRICS_PORT", "9000")

        config = Config.from_env()

        assert config.python_env == "production"
        assert config.log_level == "DEBUG"
        assert config.rate_limit == 2.5
        assert config.request_timeout == 60
        assert config.metrics_port == 9000

    def test_config_validation_success(self):
        """Test successful configuration validation."""
        config = Config(
            rate_limit=1.0,
            request_timeout=30,
            max_retries=3,
            metrics_port=8000,
            log_level="INFO",
        )

        # Should not raise
        config.validate()

    def test_config_validation_negative_rate_limit(self):
        """Test validation fails for negative rate limit."""
        config = Config(rate_limit=-1.0)

        with pytest.raises(ValueError, match="rate_limit must be non-negative"):
            config.validate()

    def test_config_validation_invalid_timeout(self):
        """Test validation fails for invalid timeout."""
        config = Config(request_timeout=0)

        with pytest.raises(ValueError, match="request_timeout must be positive"):
            config.validate()

    def test_config_validation_invalid_port(self):
        """Test validation fails for invalid port."""
        config = Config(metrics_port=99999)

        with pytest.raises(ValueError, match="metrics_port must be between"):
            config.validate()

    def test_config_validation_invalid_log_level(self):
        """Test validation fails for invalid log level."""
        config = Config(log_level="INVALID")

        with pytest.raises(ValueError, match="log_level must be one of"):
            config.validate()

    def test_is_production(self):
        """Test production environment detection."""
        config = Config(python_env="production")
        assert config.is_production() is True

        config = Config(python_env="development")
        assert config.is_production() is False

    def test_is_development(self):
        """Test development environment detection."""
        config = Config(python_env="development")
        assert config.is_development() is True

        config = Config(python_env="production")
        assert config.is_development() is False

    def test_get_config_singleton(self):
        """Test that get_config returns a singleton."""
        config1 = get_config()
        config2 = get_config()

        assert config1 is config2

    def test_reset_config(self):
        """Test configuration reset."""
        config1 = get_config()
        reset_config()
        config2 = get_config()

        assert config1 is not config2

