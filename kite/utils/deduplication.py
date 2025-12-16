"""
Case deduplication and matching utilities.
"""

from typing import List, Set, Tuple, Optional
from difflib import SequenceMatcher
import hashlib
from .data_models import CaseData
from .validation_models import ValidatedCaseData
from .logging_config import get_logger

logger = get_logger(__name__)


def normalize_case_name(case_name: str) -> str:
    """
    Normalize case name for comparison.
    
    Args:
        case_name: Original case name
        
    Returns:
        Normalized case name
    """
    # Convert to lowercase
    name = case_name.lower()
    
    # Remove common punctuation
    for char in [',', '.', ';', ':', '"', "'", '(', ')', '[', ']']:
        name = name.replace(char, '')
    
    # Normalize whitespace
    name = ' '.join(name.split())
    
    # Remove common legal terms that don't help identify uniqueness
    stopwords = ['v', 'vs', 'versus', 'et', 'al', 'and', 'the', 'in', 're']
    words = name.split()
    words = [w for w in words if w not in stopwords]
    
    return ' '.join(words)


def compute_case_hash(case_name: str, court: Optional[str], date: Optional[str]) -> str:
    """
    Compute a hash for case identification.
    
    Args:
        case_name: Case name
        court: Court name
        date: Decision date (as string)
        
    Returns:
        Hash string
    """
    # Normalize inputs
    name_norm = normalize_case_name(case_name)
    court_norm = court.lower() if court else ""
    date_norm = str(date) if date else ""
    
    # Create compound key
    key = f"{name_norm}|{court_norm}|{date_norm}"
    
    # Return SHA-256 hash
    return hashlib.sha256(key.encode('utf-8')).hexdigest()[:16]


def calculate_similarity(text1: str, text2: str) -> float:
    """
    Calculate similarity between two text strings.
    
    Args:
        text1: First text
        text2: Second text
        
    Returns:
        Similarity score between 0 and 1
    """
    return SequenceMatcher(None, text1.lower(), text2.lower()).ratio()


class CaseDeduplicator:
    """Deduplicate cases based on various matching strategies."""
    
    def __init__(self, similarity_threshold: float = 0.85):
        """
        Initialize deduplicator.
        
        Args:
            similarity_threshold: Minimum similarity to consider a match (0-1)
        """
        self.similarity_threshold = similarity_threshold
        self.logger = logger
        self._seen_hashes: Set[str] = set()
        self._seen_ids: Set[str] = set()
    
    def is_duplicate_by_hash(
        self,
        case_name: str,
        court: Optional[str] = None,
        date: Optional[str] = None
    ) -> bool:
        """
        Check if case is duplicate using hash-based matching.
        
        Args:
            case_name: Case name
            court: Court name
            date: Decision date
            
        Returns:
            True if duplicate found
        """
        case_hash = compute_case_hash(case_name, court, date)
        
        if case_hash in self._seen_hashes:
            return True
        
        self._seen_hashes.add(case_hash)
        return False
    
    def is_duplicate_by_id(self, case_id: str) -> bool:
        """
        Check if case ID has been seen.
        
        Args:
            case_id: Case identifier
            
        Returns:
            True if duplicate found
        """
        if case_id in self._seen_ids:
            return True
        
        self._seen_ids.add(case_id)
        return False
    
    def find_duplicates_in_batch(
        self,
        cases: List[CaseData]
    ) -> List[Tuple[int, int, float]]:
        """
        Find duplicate pairs in a batch of cases.
        
        Args:
            cases: List of CaseData objects
            
        Returns:
            List of tuples (index1, index2, similarity_score)
        """
        duplicates = []
        
        for i in range(len(cases)):
            for j in range(i + 1, len(cases)):
                case1 = cases[i]
                case2 = cases[j]
                
                # Quick check: exact ID match
                if case1.case_id and case2.case_id and case1.case_id == case2.case_id:
                    duplicates.append((i, j, 1.0))
                    continue
                
                # Check case name similarity
                name_sim = calculate_similarity(case1.case_name, case2.case_name)
                
                if name_sim >= self.similarity_threshold:
                    # Additional check: same court and date
                    same_court = case1.court == case2.court if case1.court and case2.court else False
                    same_date = case1.date == case2.date if case1.date and case2.date else False
                    
                    if same_court or same_date or name_sim >= 0.95:
                        duplicates.append((i, j, name_sim))
        
        self.logger.info(
            "duplicate_detection_complete",
            total_cases=len(cases),
            duplicates_found=len(duplicates),
        )
        
        return duplicates
    
    def deduplicate_batch(
        self,
        cases: List[CaseData],
        keep: str = "first"
    ) -> List[CaseData]:
        """
        Remove duplicates from a batch of cases.
        
        Args:
            cases: List of CaseData objects
            keep: Which duplicate to keep ('first' or 'last')
            
        Returns:
            List of unique cases
        """
        if not cases:
            return []
        
        # Find duplicates
        duplicate_pairs = self.find_duplicates_in_batch(cases)
        
        # Build set of indices to remove
        to_remove: Set[int] = set()
        for i, j, score in duplicate_pairs:
            if keep == "first":
                to_remove.add(j)
            else:
                to_remove.add(i)
        
        # Filter out duplicates
        unique_cases = [
            case for idx, case in enumerate(cases)
            if idx not in to_remove
        ]
        
        self.logger.info(
            "deduplication_complete",
            original_count=len(cases),
            unique_count=len(unique_cases),
            removed_count=len(to_remove),
        )
        
        return unique_cases
    
    def reset(self):
        """Reset the deduplicator state."""
        self._seen_hashes.clear()
        self._seen_ids.clear()


def deduplicate_cases(
    cases: List[CaseData],
    similarity_threshold: float = 0.85,
    keep: str = "first"
) -> List[CaseData]:
    """
    Convenience function to deduplicate a list of cases.
    
    Args:
        cases: List of CaseData objects
        similarity_threshold: Minimum similarity to consider a match (0-1)
        keep: Which duplicate to keep ('first' or 'last')
        
    Returns:
        List of unique cases
    """
    deduplicator = CaseDeduplicator(similarity_threshold=similarity_threshold)
    return deduplicator.deduplicate_batch(cases, keep=keep)
