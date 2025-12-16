"""Utility modules for Kite."""
from .helpers import setup_session, handle_request_error
from .logging_config import get_logger
from .data_models import CaseData
from .config import Config
from .exceptions import ScraperError, NetworkError, ParseError, RateLimitError
from .rate_limiter import RateLimiter
from .retry import retry
from .connection_pool import create_session
from .text_processing import clean_whitespace, remove_html_tags
from .date_parser import parse_date, format_date
from .json_utils import safe_json_loads, pretty_print_json
from .env import get_env, get_env_bool, get_env_int
from .robots_checker import RobotsChecker
from .scraping_policy import ScrapingPolicy, get_policy, get_all_policies
from .attribution import (
    Attribution,
    create_attribution,
    get_attribution_text,
    check_commercial_use_allowed,
    check_redistribution_allowed,
)

__all__ = [
    "setup_session",
    "handle_request_error",
    "get_logger",
    "CaseData",
    "Config",
    "ScraperError",
    "NetworkError",
    "ParseError",
    "RateLimitError",
    "RateLimiter",
    "retry",
    "create_session",
    "clean_whitespace",
    "remove_html_tags",
    "parse_date",
    "format_date",
    "safe_json_loads",
    "pretty_print_json",
    "get_env",
    "get_env_bool",
    "get_env_int",
    "RobotsChecker",
    "ScrapingPolicy",
    "get_policy",
    "get_all_policies",
    "Attribution",
    "create_attribution",
    "get_attribution_text",
    "check_commercial_use_allowed",
    "check_redistribution_allowed",
]
