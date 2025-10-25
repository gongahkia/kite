"""Base scraper with performance optimizations."""
from typing import Optional
from .connection_pool import create_session
from .rate_limiter import RateLimiter
from .retry import retry

class OptimizedScraperBase:
    """Base class for optimized scrapers."""
    
    def __init__(self, rate_limit: float = 1.0):
        self.session = create_session()
        self.rate_limiter = RateLimiter(rate_limit)
    
    @retry(max_attempts=3)
    def fetch(self, url: str) -> Optional[str]:
        """Fetch URL with retries and rate limiting."""
        self.rate_limiter.wait()
        response = self.session.get(url)
        response.raise_for_status()
        return response.text
    
    def close(self):
        """Close session."""
        self.session.close()
    
    def __enter__(self):
        return self
    
    def __exit__(self, *args):
        self.close()
