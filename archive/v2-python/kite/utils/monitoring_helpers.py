"""Monitoring helper utilities."""
import time
from typing import Callable, Any
from functools import wraps

def timing_decorator(func: Callable) -> Callable:
    """Decorator to measure execution time."""
    @wraps(func)
    def wrapper(*args: Any, **kwargs: Any) -> Any:
        start = time.time()
        result = func(*args, **kwargs)
        duration = time.time() - start
        print(f"{func.__name__} took {duration:.2f}s")
        return result
    return wrapper

class PerformanceMonitor:
    """Simple performance monitor."""
    
    def __init__(self):
        self.metrics = {}
    
    def record(self, name: str, value: float):
        """Record metric."""
        if name not in self.metrics:
            self.metrics[name] = []
        self.metrics[name].append(value)
    
    def get_stats(self, name: str) -> dict:
        """Get statistics for metric."""
        values = self.metrics.get(name, [])
        if not values:
            return {}
        
        return {
            "count": len(values),
            "min": min(values),
            "max": max(values),
            "avg": sum(values) / len(values),
        }
