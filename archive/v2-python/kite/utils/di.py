"""Simple dependency injection helpers."""
from typing import Dict, Any, Type, Callable

class Container:
    """Simple DI container."""
    
    def __init__(self):
        self._services: Dict[str, Any] = {}
        self._factories: Dict[str, Callable] = {}
    
    def register(self, name: str, service: Any):
        """Register service instance."""
        self._services[name] = service
    
    def register_factory(self, name: str, factory: Callable):
        """Register service factory."""
        self._factories[name] = factory
    
    def get(self, name: str) -> Any:
        """Get service instance."""
        if name in self._services:
            return self._services[name]
        
        if name in self._factories:
            service = self._factories[name]()
            self._services[name] = service
            return service
        
        raise KeyError(f"Service {name} not found")
