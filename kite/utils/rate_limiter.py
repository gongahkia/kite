"""Rate limiting utility."""
import time
from typing import Optional

class RateLimiter:
    """Simple token bucket rate limiter."""
    
    def __init__(self, rate: float = 1.0):
        self.rate = rate
        self.last_call = 0.0
    
    def wait(self) -> None:
        """Wait if necessary to respect rate limit."""
        if self.rate <= 0:
            return
        
        now = time.time()
        time_since_last = now - self.last_call
        wait_time = (1.0 / self.rate) - time_since_last
        
        if wait_time > 0:
            time.sleep(wait_time)
        
        self.last_call = time.time()
