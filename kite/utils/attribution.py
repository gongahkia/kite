"""
Attribution and data provenance helpers for ethical scraping.
"""

from typing import Optional, Dict
from dataclasses import dataclass
from datetime import datetime
from .scraping_policy import get_policy


@dataclass
class Attribution:
    """Attribution information for scraped data."""
    
    source: str
    url: str
    scraped_at: datetime
    attribution_text: Optional[str] = None
    terms_url: Optional[str] = None
    license: Optional[str] = None
    
    def to_dict(self) -> Dict:
        """Convert to dictionary."""
        return {
            "source": self.source,
            "url": self.url,
            "scraped_at": self.scraped_at.isoformat(),
            "attribution_text": self.attribution_text,
            "terms_url": self.terms_url,
            "license": self.license,
        }
    
    def format_citation(self, style: str = "text") -> str:
        """
        Format attribution as a citation.
        
        Args:
            style: Citation style ('text', 'html', 'markdown')
            
        Returns:
            Formatted citation string
        """
        if style == "html":
            parts = []
            if self.attribution_text:
                parts.append(f"<cite>{self.attribution_text}</cite>")
            parts.append(f'<a href="{self.url}">{self.url}</a>')
            parts.append(f"(accessed {self.scraped_at.strftime('%B %d, %Y')})")
            if self.terms_url:
                parts.append(f'<a href="{self.terms_url}">Terms of Service</a>')
            return ". ".join(parts)
            
        elif style == "markdown":
            parts = []
            if self.attribution_text:
                parts.append(self.attribution_text)
            parts.append(f"[{self.url}]({self.url})")
            parts.append(f"(accessed {self.scraped_at.strftime('%B %d, %Y')})")
            if self.terms_url:
                parts.append(f"[Terms of Service]({self.terms_url})")
            return ". ".join(parts)
            
        else:  # text
            parts = []
            if self.attribution_text:
                parts.append(self.attribution_text)
            parts.append(self.url)
            parts.append(f"(accessed {self.scraped_at.strftime('%B %d, %Y')})")
            if self.terms_url:
                parts.append(f"Terms: {self.terms_url}")
            return ". ".join(parts)


def create_attribution(
    jurisdiction: str,
    url: str,
    scraped_at: Optional[datetime] = None
) -> Attribution:
    """
    Create an attribution object from jurisdiction and URL.
    
    Args:
        jurisdiction: Jurisdiction identifier
        url: URL that was scraped
        scraped_at: Time of scraping (defaults to now)
        
    Returns:
        Attribution object
    """
    if scraped_at is None:
        scraped_at = datetime.now()
    
    try:
        policy = get_policy(jurisdiction)
        return Attribution(
            source=policy.jurisdiction,
            url=url,
            scraped_at=scraped_at,
            attribution_text=policy.attribution_text if policy.requires_attribution else None,
            terms_url=policy.terms_of_service_url,
            license=None,  # Could be extracted from policy
        )
    except KeyError:
        # No policy found, create basic attribution
        return Attribution(
            source=jurisdiction,
            url=url,
            scraped_at=scraped_at,
        )


def get_attribution_text(jurisdiction: str) -> Optional[str]:
    """
    Get the attribution text for a jurisdiction.
    
    Args:
        jurisdiction: Jurisdiction identifier
        
    Returns:
        Attribution text or None if not required
    """
    try:
        policy = get_policy(jurisdiction)
        return policy.attribution_text if policy.requires_attribution else None
    except KeyError:
        return None


def check_commercial_use_allowed(jurisdiction: str) -> bool:
    """
    Check if commercial use is allowed for a jurisdiction.
    
    Args:
        jurisdiction: Jurisdiction identifier
        
    Returns:
        True if commercial use is allowed
    """
    try:
        policy = get_policy(jurisdiction)
        return policy.allow_commercial_use
    except KeyError:
        # Conservative default: assume commercial use not allowed
        return False


def check_redistribution_allowed(jurisdiction: str) -> bool:
    """
    Check if redistribution is allowed for a jurisdiction.
    
    Args:
        jurisdiction: Jurisdiction identifier
        
    Returns:
        True if redistribution is allowed
    """
    try:
        policy = get_policy(jurisdiction)
        return policy.allow_redistribution
    except KeyError:
        # Conservative default: assume redistribution not allowed
        return False
