"""
Citation extraction engine for legal documents.
"""

from typing import List, Dict, Set, Optional, Tuple
from dataclasses import dataclass
import re

from .citation_patterns import (
    ALL_PATTERNS,
    CitationPattern,
    get_patterns_for_jurisdiction,
)
from .logging_config import get_logger

logger = get_logger(__name__)


@dataclass
class Citation:
    """Extracted citation with metadata."""
    
    text: str  # Original citation text
    normalized: str  # Normalized citation
    jurisdiction: str  # Jurisdiction
    pattern_name: str  # Pattern that matched
    position: int  # Character position in source text
    confidence: float = 1.0  # Confidence score (0-1)
    
    def __hash__(self):
        return hash(self.normalized)
    
    def __eq__(self, other):
        if not isinstance(other, Citation):
            return False
        return self.normalized == other.normalized


class CitationExtractor:
    """Extract and validate legal citations from text."""
    
    def __init__(self, jurisdiction: Optional[str] = None):
        """
        Initialize citation extractor.
        
        Args:
            jurisdiction: If specified, only extract citations for this jurisdiction
        """
        self.jurisdiction = jurisdiction
        self.logger = logger
        
        # Select patterns based on jurisdiction
        if jurisdiction:
            self.patterns = get_patterns_for_jurisdiction(jurisdiction)
            if not self.patterns:
                self.logger.warning(f"No patterns found for jurisdiction: {jurisdiction}")
                self.patterns = ALL_PATTERNS
        else:
            self.patterns = ALL_PATTERNS
    
    def extract(self, text: str, deduplicate: bool = True) -> List[Citation]:
        """
        Extract all citations from text.
        
        Args:
            text: Text to extract citations from
            deduplicate: Whether to remove duplicate citations
            
        Returns:
            List of extracted citations
        """
        citations = []
        
        for pattern in self.patterns:
            matches = pattern.pattern.finditer(text)
            
            for match in matches:
                citation_text = match.group(0)
                position = match.start()
                
                # Normalize citation
                normalized = self._normalize_citation(citation_text, pattern)
                
                # Create citation object
                citation = Citation(
                    text=citation_text,
                    normalized=normalized,
                    jurisdiction=pattern.jurisdiction,
                    pattern_name=pattern.name,
                    position=position,
                )
                
                citations.append(citation)
        
        # Sort by position
        citations.sort(key=lambda c: c.position)
        
        # Deduplicate if requested
        if deduplicate:
            citations = self._deduplicate(citations)
        
        self.logger.info(
            "citations_extracted",
            count=len(citations),
            jurisdiction=self.jurisdiction or "all",
        )
        
        return citations
    
    def _normalize_citation(self, text: str, pattern: CitationPattern) -> str:
        """
        Normalize a citation for comparison.
        
        Args:
            text: Original citation text
            pattern: Pattern that matched
            
        Returns:
            Normalized citation
        """
        # Basic normalization
        normalized = text.strip()
        
        # Remove extra whitespace
        normalized = ' '.join(normalized.split())
        
        # Standardize punctuation
        normalized = normalized.replace('  ', ' ')
        
        return normalized
    
    def _deduplicate(self, citations: List[Citation]) -> List[Citation]:
        """
        Remove duplicate citations.
        
        Args:
            citations: List of citations
            
        Returns:
            Deduplicated list
        """
        seen = set()
        unique = []
        
        for citation in citations:
            if citation.normalized not in seen:
                seen.add(citation.normalized)
                unique.append(citation)
        
        return unique
    
    def extract_case_names(self, text: str) -> List[str]:
        """
        Extract case names from text using common patterns.
        
        Args:
            text: Text to extract from
            
        Returns:
            List of case names
        """
        case_names = []
        
        # Pattern 1: "Party v. Party" or "Party v Party"
        pattern1 = re.compile(r'\b([A-Z][a-z]+(?:\s+[A-Z][a-z]+)*)\s+v\.?\s+([A-Z][a-z]+(?:\s+[A-Z][a-z]+)*)\b')
        
        # Pattern 2: "In re Party" or "Ex parte Party"  
        pattern2 = re.compile(r'\b(In re|Ex parte)\s+([A-Z][a-z]+(?:\s+[A-Z][a-z]+)*)\b')
        
        for match in pattern1.finditer(text):
            case_name = f"{match.group(1)} v. {match.group(2)}"
            case_names.append(case_name)
        
        for match in pattern2.finditer(text):
            case_name = f"{match.group(1)} {match.group(2)}"
            case_names.append(case_name)
        
        return case_names
    
    def validate_citation(self, citation_text: str) -> bool:
        """
        Validate if a string is a valid citation.
        
        Args:
            citation_text: Text to validate
            
        Returns:
            True if valid citation
        """
        for pattern in self.patterns:
            if pattern.pattern.search(citation_text):
                return True
        return False
    
    def get_citation_stats(self, citations: List[Citation]) -> Dict[str, any]:
        """
        Get statistics about extracted citations.
        
        Args:
            citations: List of citations
            
        Returns:
            Dictionary of statistics
        """
        if not citations:
            return {
                "total": 0,
                "by_jurisdiction": {},
                "by_pattern": {},
                "unique": 0,
            }
        
        # Count by jurisdiction
        by_jurisdiction = {}
        for citation in citations:
            by_jurisdiction[citation.jurisdiction] = by_jurisdiction.get(citation.jurisdiction, 0) + 1
        
        # Count by pattern
        by_pattern = {}
        for citation in citations:
            by_pattern[citation.pattern_name] = by_pattern.get(citation.pattern_name, 0) + 1
        
        # Count unique
        unique = len(set(c.normalized for c in citations))
        
        return {
            "total": len(citations),
            "by_jurisdiction": by_jurisdiction,
            "by_pattern": by_pattern,
            "unique": unique,
        }


def extract_citations(
    text: str,
    jurisdiction: Optional[str] = None,
    deduplicate: bool = True
) -> List[Citation]:
    """
    Convenience function to extract citations.
    
    Args:
        text: Text to extract from
        jurisdiction: Optional jurisdiction filter
        deduplicate: Whether to remove duplicates
        
    Returns:
        List of citations
    """
    extractor = CitationExtractor(jurisdiction=jurisdiction)
    return extractor.extract(text, deduplicate=deduplicate)


def extract_citation_strings(
    text: str,
    jurisdiction: Optional[str] = None,
    deduplicate: bool = True
) -> List[str]:
    """
    Convenience function to extract citation strings only.
    
    Args:
        text: Text to extract from
        jurisdiction: Optional jurisdiction filter
        deduplicate: Whether to remove duplicates
        
    Returns:
        List of citation strings
    """
    citations = extract_citations(text, jurisdiction, deduplicate)
    return [c.text for c in citations]
