"""
Citation validation and formatting utilities.
"""

from typing import Optional, List, Tuple
from dataclasses import dataclass
import re

from .citation_extractor import Citation, CitationExtractor
from .logging_config import get_logger

logger = get_logger(__name__)


@dataclass
class CitationFormat:
    """Citation format specification."""
    style: str  # bluebook, oscola, aglc, etc.
    jurisdiction: str
    
    def format(self, citation: Citation) -> str:
        """Format citation according to style."""
        # Placeholder - would implement actual formatting rules
        return citation.text


class CitationValidator:
    """Validate and verify citations."""
    
    def __init__(self):
        self.logger = logger
        self.extractor = CitationExtractor()
    
    def validate_format(self, citation: str, jurisdiction: Optional[str] = None) -> Tuple[bool, Optional[str]]:
        """
        Validate citation format.
        
        Args:
            citation: Citation string
            jurisdiction: Expected jurisdiction
            
        Returns:
            Tuple of (is_valid, error_message)
        """
        # Check if it matches any known pattern
        if not self.extractor.validate_citation(citation):
            return False, "Citation does not match any known format"
        
        # Extract and check jurisdiction if specified
        if jurisdiction:
            extracted = self.extractor.extract(citation)
            if not extracted:
                return False, "Could not parse citation"
            
            if extracted[0].jurisdiction != jurisdiction:
                return False, f"Citation is from {extracted[0].jurisdiction}, expected {jurisdiction}"
        
        return True, None
    
    def validate_completeness(self, citation: str) -> Tuple[bool, List[str]]:
        """
        Check if citation has all required components.
        
        Args:
            citation: Citation string
            
        Returns:
            Tuple of (is_complete, missing_components)
        """
        missing = []
        
        # Check for volume/year
        if not re.search(r'\d{4}|\(\d{4}\)|\[\d{4}\]', citation):
            missing.append("year")
        
        # Check for reporter/source
        if not re.search(r'[A-Z][a-z]*\.', citation):
            missing.append("reporter abbreviation")
        
        # Check for page/paragraph number
        if not re.search(r'\d+$', citation):
            missing.append("page or paragraph number")
        
        is_complete = len(missing) == 0
        return is_complete, missing
    
    def check_pinpoint(self, citation: str) -> bool:
        """
        Check if citation includes a pinpoint reference.
        
        Args:
            citation: Citation string
            
        Returns:
            True if pinpoint is present
        """
        # Look for patterns like "at 123", ", 456", " pp. 123-456"
        pinpoint_patterns = [
            r',\s*\d+',  # ", 123"
            r'\sat\s+\d+',  # "at 123"
            r'\spp?\.\s*\d+',  # "p. 123" or "pp. 123"
            r'\[?\d+\]?$',  # "[123]" at end
        ]
        
        for pattern in pinpoint_patterns:
            if re.search(pattern, citation):
                return True
        
        return False
    
    def suggest_corrections(self, citation: str) -> List[str]:
        """
        Suggest corrections for malformed citations.
        
        Args:
            citation: Possibly malformed citation
            
        Returns:
            List of suggestions
        """
        suggestions = []
        
        # Check for missing periods
        if re.search(r'[A-Z][a-z]+\s+[A-Z]', citation):
            suggestions.append("Add periods after abbreviations (e.g., 'F.3d' not 'F3d')")
        
        # Check for missing spaces
        if re.search(r'\d[A-Z]', citation):
            suggestions.append("Add space between volume and reporter")
        
        # Check for year format
        if re.search(r'\d{4}(?![)\]])', citation):
            suggestions.append("Enclose year in brackets: [2020] or (2020)")
        
        return suggestions


class CitationGraph:
    """Build and analyze citation networks."""
    
    def __init__(self):
        self.citations: Dict[str, Set[str]] = {}  # case_id -> set of cited case_ids
        self.reverse_citations: Dict[str, Set[str]] = {}  # cited_id -> set of citing case_ids
    
    def add_citation(self, citing_case: str, cited_case: str):
        """
        Add a citation relationship.
        
        Args:
            citing_case: ID of case making the citation
            cited_case: ID of case being cited
        """
        if citing_case not in self.citations:
            self.citations[citing_case] = set()
        self.citations[citing_case].add(cited_case)
        
        if cited_case not in self.reverse_citations:
            self.reverse_citations[cited_case] = set()
        self.reverse_citations[cited_case].add(citing_case)
    
    def get_cited_by(self, case_id: str) -> List[str]:
        """
        Get all cases cited by a given case.
        
        Args:
            case_id: Case identifier
            
        Returns:
            List of cited case IDs
        """
        return list(self.citations.get(case_id, set()))
    
    def get_citing(self, case_id: str) -> List[str]:
        """
        Get all cases that cite a given case.
        
        Args:
            case_id: Case identifier
            
        Returns:
            List of citing case IDs
        """
        return list(self.reverse_citations.get(case_id, set()))
    
    def get_citation_count(self, case_id: str) -> int:
        """
        Get number of times a case has been cited.
        
        Args:
            case_id: Case identifier
            
        Returns:
            Citation count
        """
        return len(self.reverse_citations.get(case_id, set()))
    
    def get_most_cited(self, limit: int = 10) -> List[Tuple[str, int]]:
        """
        Get most frequently cited cases.
        
        Args:
            limit: Maximum number to return
            
        Returns:
            List of (case_id, citation_count) tuples
        """
        citation_counts = [
            (case_id, len(citing_cases))
            for case_id, citing_cases in self.reverse_citations.items()
        ]
        
        citation_counts.sort(key=lambda x: x[1], reverse=True)
        return citation_counts[:limit]
    
    def find_common_citations(self, case_id1: str, case_id2: str) -> List[str]:
        """
        Find cases cited by both cases.
        
        Args:
            case_id1: First case ID
            case_id2: Second case ID
            
        Returns:
            List of commonly cited case IDs
        """
        cited1 = set(self.get_cited_by(case_id1))
        cited2 = set(self.get_cited_by(case_id2))
        return list(cited1.intersection(cited2))


def validate_citation(citation: str, jurisdiction: Optional[str] = None) -> Tuple[bool, Optional[str]]:
    """
    Convenience function to validate a citation.
    
    Args:
        citation: Citation string
        jurisdiction: Expected jurisdiction
        
    Returns:
        Tuple of (is_valid, error_message)
    """
    validator = CitationValidator()
    return validator.validate_format(citation, jurisdiction)
