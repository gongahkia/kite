"""Simple in-memory cache implementation."""
from typing import Any, Optional
import time

class SimpleCache:
    """Thread-unsafe simple cache with TTL."""
    
    def __init__(self, ttl: int = 300):
        self._cache = {}
        self._ttl = ttl
    
    def get(self, key: str) -> Optional[Any]:
        """Get value from cache."""
        if key in self._cache:
            value, timestamp = self._cache[key]
            if time.time() - timestamp < self._ttl:
                return value
            del self._cache[key]
        return None
    
    def set(self, key: str, value: Any) -> None:
        """Set value in cache."""
        self._cache[key] = (value, time.time())
    
    def clear(self) -> None:
        """Clear cache."""
        self._cache.clear()
