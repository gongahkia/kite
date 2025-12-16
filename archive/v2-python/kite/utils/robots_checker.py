"""
Robots.txt parser and compliance checker for ethical web scraping.
"""

import urllib.robotparser
from typing import Optional, Dict
from urllib.parse import urlparse
import time
from .logging_config import get_logger

logger = get_logger(__name__)


class RobotsChecker:
    """Check robots.txt compliance for URLs."""
    
    def __init__(self, user_agent: str = "KiteBot"):
        """
        Initialize the robots.txt checker.
        
        Args:
            user_agent: User agent string to identify the bot
        """
        self.user_agent = user_agent
        self._cache: Dict[str, urllib.robotparser.RobotFileParser] = {}
        self._cache_timestamps: Dict[str, float] = {}
        self._cache_ttl = 3600  # Cache for 1 hour
        self.logger = logger
    
    def _get_robots_url(self, url: str) -> str:
        """Get the robots.txt URL for a given URL."""
        parsed = urlparse(url)
        return f"{parsed.scheme}://{parsed.netloc}/robots.txt"
    
    def _get_parser(self, url: str) -> Optional[urllib.robotparser.RobotFileParser]:
        """
        Get or create a RobotFileParser for the domain.
        
        Args:
            url: URL to check
            
        Returns:
            RobotFileParser instance or None if robots.txt cannot be fetched
        """
        robots_url = self._get_robots_url(url)
        current_time = time.time()
        
        # Check cache
        if robots_url in self._cache:
            if current_time - self._cache_timestamps.get(robots_url, 0) < self._cache_ttl:
                return self._cache[robots_url]
        
        # Fetch and parse robots.txt
        try:
            parser = urllib.robotparser.RobotFileParser()
            parser.set_url(robots_url)
            parser.read()
            
            # Cache the parser
            self._cache[robots_url] = parser
            self._cache_timestamps[robots_url] = current_time
            
            self.logger.info(f"Loaded robots.txt from {robots_url}")
            return parser
            
        except Exception as e:
            self.logger.warning(f"Could not fetch robots.txt from {robots_url}: {e}")
            return None
    
    def can_fetch(self, url: str) -> bool:
        """
        Check if the URL can be fetched according to robots.txt.
        
        Args:
            url: URL to check
            
        Returns:
            True if URL can be fetched, False otherwise
        """
        parser = self._get_parser(url)
        
        if parser is None:
            # If robots.txt cannot be fetched, assume we can fetch
            # (conservative approach - some sites don't have robots.txt)
            self.logger.debug(f"No robots.txt found for {url}, assuming allowed")
            return True
        
        can_fetch = parser.can_fetch(self.user_agent, url)
        
        if not can_fetch:
            self.logger.warning(f"robots.txt blocks access to {url} for {self.user_agent}")
        
        return can_fetch
    
    def get_crawl_delay(self, url: str) -> Optional[float]:
        """
        Get the crawl delay specified in robots.txt.
        
        Args:
            url: URL to check
            
        Returns:
            Crawl delay in seconds, or None if not specified
        """
        parser = self._get_parser(url)
        
        if parser is None:
            return None
        
        delay = parser.crawl_delay(self.user_agent)
        
        if delay:
            self.logger.info(f"Crawl delay for {url}: {delay} seconds")
        
        return delay
    
    def clear_cache(self):
        """Clear the robots.txt cache."""
        self._cache.clear()
        self._cache_timestamps.clear()
        self.logger.info("Cleared robots.txt cache")
