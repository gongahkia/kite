"""
Scraping policies and terms of service compliance configuration.
"""

from typing import Dict, Optional
from dataclasses import dataclass


@dataclass
class ScrapingPolicy:
    """Policy configuration for scraping a jurisdiction."""
    
    jurisdiction: str
    base_url: str
    
    # Rate limiting
    rate_limit: float = 1.0  # seconds between requests
    max_concurrent_requests: int = 1
    
    # Retry behavior
    max_retries: int = 3
    retry_delay: float = 1.0
    backoff_factor: float = 2.0
    
    # Timeouts
    request_timeout: int = 30
    
    # Politeness
    respect_robots_txt: bool = True
    user_agent: Optional[str] = None
    
    # Legal/Terms of Service
    requires_attribution: bool = False
    attribution_text: Optional[str] = None
    terms_of_service_url: Optional[str] = None
    
    # Access restrictions
    requires_authentication: bool = False
    api_key_required: bool = False
    
    # Data usage
    allow_commercial_use: bool = True
    allow_redistribution: bool = True
    
    # Notes
    notes: Optional[str] = None


# Jurisdiction-specific scraping policies
JURISDICTION_POLICIES: Dict[str, ScrapingPolicy] = {
    "courtlistener": ScrapingPolicy(
        jurisdiction="US (Federal and State)",
        base_url="https://www.courtlistener.com",
        rate_limit=1.0,
        respect_robots_txt=True,
        requires_attribution=True,
        attribution_text="Data from Free Law Project's CourtListener",
        terms_of_service_url="https://www.courtlistener.com/terms/",
        allow_commercial_use=True,
        allow_redistribution=True,
        notes="Part of Free Law Project - open legal data initiative"
    ),
    
    "canlii": ScrapingPolicy(
        jurisdiction="Canada",
        base_url="https://www.canlii.org",
        rate_limit=2.0,  # Be more conservative
        respect_robots_txt=True,
        requires_attribution=True,
        attribution_text="Data from CanLII (Canadian Legal Information Institute)",
        terms_of_service_url="https://www.canlii.org/en/info/terms.html",
        allow_commercial_use=True,
        allow_redistribution=True,
        notes="Nonprofit providing free access to Canadian law"
    ),
    
    "bailii": ScrapingPolicy(
        jurisdiction="UK and Ireland",
        base_url="https://www.bailii.org",
        rate_limit=2.0,
        respect_robots_txt=True,
        requires_attribution=True,
        attribution_text="Data from BAILII (British and Irish Legal Information Institute)",
        terms_of_service_url="https://www.bailii.org/bailii/info.html",
        allow_commercial_use=True,
        allow_redistribution=True,
        notes="Free access to British and Irish case law and legislation"
    ),
    
    "austlii": ScrapingPolicy(
        jurisdiction="Australia",
        base_url="https://www.austlii.edu.au",
        rate_limit=2.0,
        respect_robots_txt=True,
        requires_attribution=True,
        attribution_text="Data from AustLII (Australasian Legal Information Institute)",
        terms_of_service_url="https://www.austlii.edu.au/austlii/info/terms.html",
        allow_commercial_use=True,
        allow_redistribution=True,
        notes="Free access to Australian legal materials"
    ),
    
    "worldlii": ScrapingPolicy(
        jurisdiction="International",
        base_url="https://www.worldlii.org",
        rate_limit=2.0,
        respect_robots_txt=True,
        requires_attribution=True,
        attribution_text="Data from WorldLII (World Legal Information Institute)",
        terms_of_service_url="https://www.worldlii.org/worldlii/info/terms.html",
        allow_commercial_use=True,
        allow_redistribution=True,
        notes="Global free access to law"
    ),
    
    "hklii": ScrapingPolicy(
        jurisdiction="Hong Kong",
        base_url="https://www.hklii.hk",
        rate_limit=2.0,
        respect_robots_txt=True,
        requires_attribution=True,
        attribution_text="Data from HKLII (Hong Kong Legal Information Institute)",
        allow_commercial_use=True,
        allow_redistribution=True,
    ),
    
    "singapore": ScrapingPolicy(
        jurisdiction="Singapore",
        base_url="https://www.judiciary.gov.sg",
        rate_limit=3.0,  # Government site - be more conservative
        max_retries=2,
        respect_robots_txt=True,
        requires_attribution=True,
        attribution_text="Data from Singapore Courts",
        terms_of_service_url="https://www.judiciary.gov.sg/terms-use",
        allow_commercial_use=False,  # Check terms carefully
        notes="Official Singapore judiciary website"
    ),
    
    "indian-kanoon": ScrapingPolicy(
        jurisdiction="India",
        base_url="https://indiankanoon.org",
        rate_limit=2.0,
        respect_robots_txt=True,
        requires_attribution=True,
        attribution_text="Data from Indian Kanoon",
        allow_commercial_use=True,
        allow_redistribution=True,
        notes="Free access to Indian case law"
    ),
    
    "supremecourt-india": ScrapingPolicy(
        jurisdiction="India (Supreme Court)",
        base_url="https://main.sci.gov.in",
        rate_limit=3.0,  # Government site
        max_retries=2,
        respect_robots_txt=True,
        requires_attribution=True,
        attribution_text="Data from Supreme Court of India",
        notes="Official Supreme Court of India website"
    ),
    
    "legifrance": ScrapingPolicy(
        jurisdiction="France",
        base_url="https://www.legifrance.gouv.fr",
        rate_limit=3.0,  # Government site
        max_retries=2,
        respect_robots_txt=True,
        requires_attribution=True,
        attribution_text="Data from Légifrance",
        terms_of_service_url="https://www.legifrance.gouv.fr/Informations/Informations-legales",
        notes="Official French legal database"
    ),
    
    "german-law": ScrapingPolicy(
        jurisdiction="Germany",
        base_url="https://www.gesetze-im-internet.de",
        rate_limit=3.0,
        respect_robots_txt=True,
        requires_attribution=True,
        attribution_text="Data from Gesetze im Internet",
        notes="German federal law archive"
    ),
    
    "curia-europa": ScrapingPolicy(
        jurisdiction="European Union",
        base_url="https://curia.europa.eu",
        rate_limit=3.0,  # EU official site
        max_retries=2,
        respect_robots_txt=True,
        requires_attribution=True,
        attribution_text="Data from Court of Justice of the European Union",
        notes="Official CJEU case law"
    ),
    
    "kenya-law": ScrapingPolicy(
        jurisdiction="Kenya",
        base_url="http://kenyalaw.org",
        rate_limit=2.0,
        respect_robots_txt=True,
        requires_attribution=True,
        attribution_text="Data from Kenya Law Reports",
        notes="Kenya legal information platform"
    ),
    
    "supremecourt-japan": ScrapingPolicy(
        jurisdiction="Japan (Supreme Court)",
        base_url="https://www.courts.go.jp",
        rate_limit=3.0,
        respect_robots_txt=True,
        requires_attribution=True,
        attribution_text="Data from Supreme Court of Japan",
        notes="Official Japanese court website"
    ),
    
    "worldcourts": ScrapingPolicy(
        jurisdiction="International Courts",
        base_url="https://www.worldcourts.com",
        rate_limit=2.0,
        respect_robots_txt=True,
        allow_commercial_use=True,
        notes="International and regional court decisions"
    ),
    
    "legal-tools": ScrapingPolicy(
        jurisdiction="International Criminal Law",
        base_url="https://www.legal-tools.org",
        rate_limit=2.0,
        respect_robots_txt=True,
        requires_attribution=True,
        allow_commercial_use=True,
        notes="International criminal law database"
    ),
    
    "findlaw": ScrapingPolicy(
        jurisdiction="US",
        base_url="https://caselaw.findlaw.com",
        rate_limit=2.0,
        respect_robots_txt=True,
        requires_attribution=False,
        allow_commercial_use=False,  # Commercial site - check terms
        notes="Thomson Reuters FindLaw - commercial service"
    ),
}


def get_policy(jurisdiction: str) -> ScrapingPolicy:
    """
    Get the scraping policy for a jurisdiction.
    
    Args:
        jurisdiction: Jurisdiction identifier
        
    Returns:
        ScrapingPolicy for the jurisdiction
        
    Raises:
        KeyError: If jurisdiction not found
    """
    if jurisdiction not in JURISDICTION_POLICIES:
        raise KeyError(f"No scraping policy defined for jurisdiction: {jurisdiction}")
    
    return JURISDICTION_POLICIES[jurisdiction]


def get_all_policies() -> Dict[str, ScrapingPolicy]:
    """Get all defined scraping policies."""
    return JURISDICTION_POLICIES.copy()
