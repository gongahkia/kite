"""Environment variable utilities."""
import os
from typing import Optional, TypeVar, Callable

T = TypeVar('T')

def get_env(key: str, default: Optional[str] = None) -> Optional[str]:
    """Get environment variable with default."""
    return os.getenv(key, default)

def get_env_bool(key: str, default: bool = False) -> bool:
    """Get boolean environment variable."""
    value = os.getenv(key)
    if value is None:
        return default
    return value.lower() in ('true', '1', 'yes', 'on')

def get_env_int(key: str, default: int = 0) -> int:
    """Get integer environment variable."""
    try:
        return int(os.getenv(key, default))
    except (ValueError, TypeError):
        return default

def get_env_float(key: str, default: float = 0.0) -> float:
    """Get float environment variable."""
    try:
        return float(os.getenv(key, default))
    except (ValueError, TypeError):
        return default

def require_env(key: str) -> str:
    """Get required environment variable or raise error."""
    value = os.getenv(key)
    if value is None:
        raise ValueError(f"Required environment variable {key} is not set")
    return value
