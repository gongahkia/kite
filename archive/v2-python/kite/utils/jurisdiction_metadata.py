"""
Jurisdiction-specific metadata and enrichment utilities.
"""

from typing import Dict, Optional, List
from enum import Enum
from dataclasses import dataclass


class CourtType(str, Enum):
    """Type of court."""
    SUPREME = "supreme"
    APPELLATE = "appellate"
    TRIAL = "trial"
    DISTRICT = "district"
    MAGISTRATE = "magistrate"
    TRIBUNAL = "tribunal"
    ADMINISTRATIVE = "administrative"
    CONSTITUTIONAL = "constitutional"
    INTERNATIONAL = "international"


class CaseType(str, Enum):
    """Type of legal case."""
    CIVIL = "civil"
    CRIMINAL = "criminal"
    CONSTITUTIONAL = "constitutional"
    ADMINISTRATIVE = "administrative"
    FAMILY = "family"
    COMMERCIAL = "commercial"
    LABOR = "labor"
    TAX = "tax"
    IMMIGRATION = "immigration"
    INTELLECTUAL_PROPERTY = "intellectual_property"


@dataclass
class JurisdictionMetadata:
    """Metadata specific to a legal jurisdiction."""
    
    # Court hierarchy
    court_level: int = 0  # 0=unknown, 1=lowest, higher numbers=higher courts
    court_type: Optional[CourtType] = None
    court_division: Optional[str] = None  # e.g., "Civil Division", "Criminal Division"
    
    # Case categorization
    case_type: Optional[CaseType] = None
    procedural_posture: Optional[str] = None  # e.g., "appeal", "motion", "trial"
    disposition: Optional[str] = None  # e.g., "affirmed", "reversed", "remanded"
    
    # Authority and precedent
    precedential_status: str = "unknown"  # "binding", "persuasive", "non-precedential"
    is_published: bool = True
    is_reported: bool = True
    
    # Panel information
    panel_size: Optional[int] = None
    is_en_banc: bool = False
    is_per_curiam: bool = False
    
    # Procedural information
    has_dissent: bool = False
    has_concurrence: bool = False
    is_unanimous: bool = False
    
    # Case history
    prior_history: List[str] = None
    subsequent_history: List[str] = None
    related_cases: List[str] = None
    
    def __post_init__(self):
        if self.prior_history is None:
            self.prior_history = []
        if self.subsequent_history is None:
            self.subsequent_history = []
        if self.related_cases is None:
            self.related_cases = []


# Court hierarchy definitions by jurisdiction
COURT_HIERARCHIES = {
    "US": {
        "Supreme Court of the United States": 5,
        "U.S. Supreme Court": 5,
        "United States Court of Appeals": 4,
        "U.S. Court of Appeals": 4,
        "United States District Court": 3,
        "U.S. District Court": 3,
        "State Supreme Court": 4,
        "State Appellate Court": 3,
        "State Trial Court": 2,
    },
    "UK": {
        "Supreme Court": 5,
        "Court of Appeal": 4,
        "High Court": 3,
        "Crown Court": 2,
        "County Court": 2,
        "Magistrates' Court": 1,
    },
    "Canada": {
        "Supreme Court of Canada": 5,
        "Federal Court of Appeal": 4,
        "Provincial Court of Appeal": 4,
        "Federal Court": 3,
        "Superior Court": 3,
        "Provincial Court": 2,
    },
    "Australia": {
        "High Court of Australia": 5,
        "Federal Court of Australia": 4,
        "State Supreme Court": 4,
        "District Court": 3,
        "Magistrates Court": 2,
    },
}


def determine_court_level(court_name: str, jurisdiction: str) -> int:
    """
    Determine the hierarchical level of a court.
    
    Args:
        court_name: Name of the court
        jurisdiction: Jurisdiction identifier
        
    Returns:
        Court level (1-5, 0 if unknown)
    """
    if jurisdiction not in COURT_HIERARCHIES:
        return 0
    
    hierarchy = COURT_HIERARCHIES[jurisdiction]
    
    # Try exact match first
    if court_name in hierarchy:
        return hierarchy[court_name]
    
    # Try partial matching
    court_lower = court_name.lower()
    for court, level in hierarchy.items():
        if court.lower() in court_lower:
            return level
    
    # Try keyword matching
    if "supreme" in court_lower:
        return 5
    elif "appeal" in court_lower or "appellate" in court_lower:
        return 4
    elif "high" in court_lower or "superior" in court_lower:
        return 3
    elif "district" in court_lower or "trial" in court_lower:
        return 2
    elif "magistrate" in court_lower:
        return 1
    
    return 0


def determine_court_type(court_name: str) -> Optional[CourtType]:
    """
    Determine the type of court from its name.
    
    Args:
        court_name: Name of the court
        
    Returns:
        CourtType or None
    """
    court_lower = court_name.lower()
    
    if "supreme" in court_lower:
        return CourtType.SUPREME
    elif "appeal" in court_lower or "appellate" in court_lower:
        return CourtType.APPELLATE
    elif "trial" in court_lower:
        return CourtType.TRIAL
    elif "district" in court_lower:
        return CourtType.DISTRICT
    elif "magistrate" in court_lower:
        return CourtType.MAGISTRATE
    elif "tribunal" in court_lower:
        return CourtType.TRIBUNAL
    elif "administrative" in court_lower:
        return CourtType.ADMINISTRATIVE
    elif "constitutional" in court_lower:
        return CourtType.CONSTITUTIONAL
    elif "international" in court_lower:
        return CourtType.INTERNATIONAL
    
    return None


def extract_court_division(court_name: str) -> Optional[str]:
    """
    Extract court division from court name.
    
    Args:
        court_name: Full court name
        
    Returns:
        Division name or None
    """
    import re
    
    # Common division patterns
    patterns = [
        r'\((.*?Division)\)',
        r'-(.*?Division)',
        r',(.*?Division)',
        r'\((Civil|Criminal|Family|Commercial|Chancery)\)',
    ]
    
    for pattern in patterns:
        match = re.search(pattern, court_name, re.IGNORECASE)
        if match:
            return match.group(1).strip()
    
    return None


def determine_case_type(case_name: str, summary: Optional[str] = None) -> Optional[CaseType]:
    """
    Determine case type from case name and summary.
    
    Args:
        case_name: Case name
        summary: Optional case summary
        
    Returns:
        CaseType or None
    """
    text = (case_name + " " + (summary or "")).lower()
    
    # Criminal keywords
    criminal_keywords = ["criminal", "prosecution", "defendant", "guilty", "sentencing", 
                        "conviction", "murder", "theft", "assault", "drug", "dui"]
    
    # Civil keywords
    civil_keywords = ["damages", "negligence", "breach of contract", "tort", "plaintiff",
                     "defendant sued", "civil action"]
    
    # Family keywords
    family_keywords = ["divorce", "custody", "adoption", "marriage", "domestic"]
    
    # Count keyword matches
    criminal_score = sum(1 for kw in criminal_keywords if kw in text)
    civil_score = sum(1 for kw in civil_keywords if kw in text)
    family_score = sum(1 for kw in family_keywords if kw in text)
    
    scores = {
        CaseType.CRIMINAL: criminal_score,
        CaseType.CIVIL: civil_score,
        CaseType.FAMILY: family_score,
    }
    
    max_score = max(scores.values())
    if max_score == 0:
        return None
    
    for case_type, score in scores.items():
        if score == max_score:
            return case_type
    
    return None


def enrich_case_metadata(
    court_name: Optional[str],
    jurisdiction: str,
    case_name: Optional[str] = None,
    summary: Optional[str] = None,
) -> JurisdictionMetadata:
    """
    Enrich case with jurisdiction-specific metadata.
    
    Args:
        court_name: Name of the court
        jurisdiction: Jurisdiction identifier
        case_name: Optional case name
        summary: Optional case summary
        
    Returns:
        JurisdictionMetadata object
    """
    metadata = JurisdictionMetadata()
    
    if court_name:
        metadata.court_level = determine_court_level(court_name, jurisdiction)
        metadata.court_type = determine_court_type(court_name)
        metadata.court_division = extract_court_division(court_name)
    
    if case_name:
        metadata.case_type = determine_case_type(case_name, summary)
    
    return metadata
