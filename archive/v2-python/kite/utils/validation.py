"""Input validation utilities."""
from typing import Any, Optional
import re

def validate_string(value: Any, min_length: int = 0, max_length: Optional[int] = None) -> bool:
    """Validate string length."""
    if not isinstance(value, str):
        return False
    if len(value) < min_length:
        return False
    if max_length and len(value) > max_length:
        return False
    return True

def validate_email(email: str) -> bool:
    """Validate email format."""
    pattern = r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
    return bool(re.match(pattern, email))

def validate_range(value: Any, min_val: float, max_val: float) -> bool:
    """Validate numeric range."""
    try:
        num = float(value)
        return min_val <= num <= max_val
    except (ValueError, TypeError):
        return False
