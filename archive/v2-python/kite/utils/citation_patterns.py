"""
Legal citation patterns and formats for various jurisdictions.
"""

import re
from typing import Dict, List, Pattern
from dataclasses import dataclass


@dataclass
class CitationPattern:
    """Pattern for matching legal citations."""
    name: str
    pattern: Pattern
    jurisdiction: str
    format_example: str


# US Bluebook Citation Patterns
US_PATTERNS = [
    CitationPattern(
        name="US Supreme Court Reporter",
        pattern=re.compile(r'\b(\d+)\s+U\.?S\.?\s+(\d+)\b', re.IGNORECASE),
        jurisdiction="US",
        format_example="123 U.S. 456"
    ),
    CitationPattern(
        name="Federal Reporter (F., F.2d, F.3d, F.4th)",
        pattern=re.compile(r'\b(\d+)\s+F\.(?:(2d|3d|4th))?\s+(\d+)\b', re.IGNORECASE),
        jurisdiction="US",
        format_example="123 F.3d 456"
    ),
    CitationPattern(
        name="Federal Supplement (F.Supp., F.Supp.2d, F.Supp.3d)",
        pattern=re.compile(r'\b(\d+)\s+F\.\s?Supp\.(?:(2d|3d))?\s+(\d+)\b', re.IGNORECASE),
        jurisdiction="US",
        format_example="123 F.Supp.2d 456"
    ),
    CitationPattern(
        name="Supreme Court Reporter",
        pattern=re.compile(r'\b(\d+)\s+S\.?\s?Ct\.?\s+(\d+)\b', re.IGNORECASE),
        jurisdiction="US",
        format_example="123 S.Ct. 456"
    ),
    CitationPattern(
        name="Lawyers Edition (L.Ed., L.Ed.2d)",
        pattern=re.compile(r'\b(\d+)\s+L\.?\s?Ed\.?(2d)?\s+(\d+)\b', re.IGNORECASE),
        jurisdiction="US",
        format_example="123 L.Ed.2d 456"
    ),
    CitationPattern(
        name="State Reporter",
        pattern=re.compile(r'\b(\d+)\s+([A-Z][a-z]{1,4}\.(?:2d|3d)?)\s+(\d+)\b'),
        jurisdiction="US",
        format_example="123 Cal.2d 456"
    ),
]

# Canadian Citation Patterns  
CANADIAN_PATTERNS = [
    CitationPattern(
        name="Supreme Court Reports (SCR)",
        pattern=re.compile(r'\[(\d{4})\]\s+(\d+)\s+S\.?C\.?R\.?\s+(\d+)', re.IGNORECASE),
        jurisdiction="Canada",
        format_example="[2020] 1 S.C.R. 123"
    ),
    CitationPattern(
        name="Federal Court Reports",
        pattern=re.compile(r'\[(\d{4})\]\s+(\d+)\s+F\.?C\.?R\.?\s+(\d+)', re.IGNORECASE),
        jurisdiction="Canada",
        format_example="[2020] 1 F.C.R. 123"
    ),
    CitationPattern(
        name="CanLII Citation",
        pattern=re.compile(r'(\d{4})\s+([A-Z]{2,5})\s+(\d+)', re.IGNORECASE),
        jurisdiction="Canada",
        format_example="2020 SCC 15"
    ),
]

# UK Citation Patterns
UK_PATTERNS = [
    CitationPattern(
        name="Neutral Citation (UK)",
        pattern=re.compile(r'\[(\d{4})\]\s+(UKSC|UKHL|EWCA|EWHC|UKUT|UKFTT)\s+(Civ|Crim|Admin|Ch|Comm|Fam|QB|Pat)?\s*(\d+)', re.IGNORECASE),
        jurisdiction="UK",
        format_example="[2020] UKSC 15"
    ),
    CitationPattern(
        name="Appeal Cases (AC)",
        pattern=re.compile(r'\[(\d{4})\]\s+(AC)\s+(\d+)', re.IGNORECASE),
        jurisdiction="UK",
        format_example="[2020] AC 123"
    ),
    CitationPattern(
        name="Weekly Law Reports (WLR)",
        pattern=re.compile(r'\[(\d{4})\]\s+(\d+)\s+W\.?L\.?R\.?\s+(\d+)', re.IGNORECASE),
        jurisdiction="UK",
        format_example="[2020] 1 WLR 123"
    ),
    CitationPattern(
        name="All England Reports (All ER)",
        pattern=re.compile(r'\[(\d{4})\]\s+(\d+)\s+All\s+E\.?R\.?\s+(\d+)', re.IGNORECASE),
        jurisdiction="UK",
        format_example="[2020] 1 All ER 123"
    ),
]

# Australian Citation Patterns
AUSTRALIAN_PATTERNS = [
    CitationPattern(
        name="Commonwealth Law Reports (CLR)",
        pattern=re.compile(r'\((\d{4})\)\s+(\d+)\s+C\.?L\.?R\.?\s+(\d+)', re.IGNORECASE),
        jurisdiction="Australia",
        format_example="(2020) 123 CLR 456"
    ),
    CitationPattern(
        name="Federal Court Reports (FCR)",
        pattern=re.compile(r'\((\d{4})\)\s+(\d+)\s+F\.?C\.?R\.?\s+(\d+)', re.IGNORECASE),
        jurisdiction="Australia",
        format_example="(2020) 123 FCR 456"
    ),
    CitationPattern(
        name="AustLII Medium Neutral Citation",
        pattern=re.compile(r'\[(\d{4})\]\s+([A-Z]{3,})\s+(\d+)', re.IGNORECASE),
        jurisdiction="Australia",
        format_example="[2020] HCA 15"
    ),
]

# European Citation Patterns
EU_PATTERNS = [
    CitationPattern(
        name="ECLI Citation",
        pattern=re.compile(r'ECLI:[A-Z]{2}:[A-Z]+:\d{4}:\d+', re.IGNORECASE),
        jurisdiction="EU",
        format_example="ECLI:EU:C:2020:123"
    ),
    CitationPattern(
        name="EU Case Number",
        pattern=re.compile(r'Case\s+C-(\d+/\d+)', re.IGNORECASE),
        jurisdiction="EU",
        format_example="Case C-123/20"
    ),
]

# International Citation Patterns
INTERNATIONAL_PATTERNS = [
    CitationPattern(
        name="ICJ Reports",
        pattern=re.compile(r'I\.?C\.?J\.?\s+Reports\s+\(?\d{4}\)?,?\s+(\d+)', re.IGNORECASE),
        jurisdiction="International",
        format_example="ICJ Reports (2020), 123"
    ),
    CitationPattern(
        name="ECHR Citation",
        pattern=re.compile(r'ECHR\s+\d{4}-[IVX]+', re.IGNORECASE),
        jurisdiction="International",
        format_example="ECHR 2020-III"
    ),
]

# All patterns combined
ALL_PATTERNS: List[CitationPattern] = (
    US_PATTERNS +
    CANADIAN_PATTERNS +
    UK_PATTERNS +
    AUSTRALIAN_PATTERNS +
    EU_PATTERNS +
    INTERNATIONAL_PATTERNS
)

# Pattern lookup by jurisdiction
PATTERNS_BY_JURISDICTION: Dict[str, List[CitationPattern]] = {
    "US": US_PATTERNS,
    "Canada": CANADIAN_PATTERNS,
    "UK": UK_PATTERNS,
    "Australia": AUSTRALIAN_PATTERNS,
    "EU": EU_PATTERNS,
    "International": INTERNATIONAL_PATTERNS,
}


def get_patterns_for_jurisdiction(jurisdiction: str) -> List[CitationPattern]:
    """
    Get citation patterns for a specific jurisdiction.
    
    Args:
        jurisdiction: Jurisdiction identifier
        
    Returns:
        List of citation patterns
    """
    return PATTERNS_BY_JURISDICTION.get(jurisdiction, [])
