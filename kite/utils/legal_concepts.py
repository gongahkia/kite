"""
Legal concept extraction and tagging system.
"""

from typing import List, Dict, Set, Tuple
from dataclasses import dataclass
from collections import Counter
import re

from .logging_config import get_logger

logger = get_logger(__name__)


@dataclass
class LegalConcept:
    """Extracted legal concept with metadata."""
    concept: str
    category: str
    confidence: float
    positions: List[int]  # Character positions where found
    
    def __hash__(self):
        return hash((self.concept, self.category))
    
    def __eq__(self, other):
        if not isinstance(other, LegalConcept):
            return False
        return self.concept == other.concept and self.category == other.category


# Legal concept taxonomy
LEGAL_CONCEPTS = {
    "constitutional_law": [
        "due process", "equal protection", "first amendment", "second amendment",
        "fourth amendment", "fifth amendment", "freedom of speech", "freedom of religion",
        "right to bear arms", "search and seizure", "cruel and unusual punishment",
        "double jeopardy", "self-incrimination", "substantive due process",
        "procedural due process", "strict scrutiny", "intermediate scrutiny",
        "rational basis", "commerce clause", "supremacy clause",
    ],
    
    "criminal_law": [
        "mens rea", "actus reus", "intent", "negligence", "recklessness",
        "causation", "burden of proof", "beyond reasonable doubt",
        "probable cause", "miranda rights", "right to counsel",
        "speedy trial", "jury trial", "plea bargain", "sentencing guidelines",
        "aggravating factors", "mitigating factors", "habeas corpus",
    ],
    
    "contract_law": [
        "offer", "acceptance", "consideration", "mutual assent",
        "breach of contract", "material breach", "anticipatory repudiation",
        "specific performance", "damages", "consequential damages",
        "liquidated damages", "statute of frauds", "parol evidence rule",
        "impossibility", "impracticability", "frustration of purpose",
        "good faith", "unconscionability",
    ],
    
    "tort_law": [
        "negligence", "duty of care", "breach of duty", "causation",
        "proximate cause", "damages", "strict liability", "res ipsa loquitur",
        "comparative negligence", "contributory negligence", "assumption of risk",
        "intentional tort", "battery", "assault", "false imprisonment",
        "defamation", "libel", "slander", "invasion of privacy",
        "emotional distress", "punitive damages",
    ],
    
    "property_law": [
        "real property", "personal property", "title", "deed",
        "easement", "covenant", "adverse possession", "eminent domain",
        "zoning", "nuisance", "trespass", "landlord-tenant",
        "lease", "eviction", "quiet enjoyment", "habitability",
    ],
    
    "civil_procedure": [
        "jurisdiction", "venue", "standing", "justiciability",
        "mootness", "ripeness", "class action", "summary judgment",
        "motion to dismiss", "discovery", "deposition", "interrogatories",
        "appeal", "standard of review", "de novo review", "abuse of discretion",
        "clearly erroneous", "substantial evidence",
    ],
    
    "evidence": [
        "hearsay", "relevance", "materiality", "prejudicial",
        "authentication", "best evidence rule", "privilege",
        "attorney-client privilege", "work product", "expert testimony",
        "lay witness", "impeachment", "credibility", "burden of proof",
        "preponderance of evidence",
    ],
    
    "administrative_law": [
        "agency action", "rulemaking", "adjudication", "arbitrary and capricious",
        "chevron deference", "auer deference", "exhaustion of remedies",
        "primary jurisdiction", "ripeness", "standing",
    ],
    
    "corporate_law": [
        "fiduciary duty", "duty of care", "duty of loyalty", "business judgment rule",
        "piercing the corporate veil", "shareholder derivative suit",
        "merger", "acquisition", "securities", "disclosure",
    ],
    
    "intellectual_property": [
        "patent", "trademark", "copyright", "trade secret",
        "infringement", "fair use", "likelihood of confusion",
        "substantial similarity", "obviousness", "novelty",
    ],
}

# Legal standards and tests
LEGAL_STANDARDS = {
    "standards_of_review": [
        "de novo", "clearly erroneous", "abuse of discretion",
        "arbitrary and capricious", "substantial evidence",
    ],
    
    "burdens_of_proof": [
        "beyond reasonable doubt", "clear and convincing evidence",
        "preponderance of the evidence", "probable cause",
        "reasonable suspicion",
    ],
    
    "scrutiny_levels": [
        "strict scrutiny", "intermediate scrutiny", "rational basis review",
        "heightened scrutiny",
    ],
}

# Causes of action
CAUSES_OF_ACTION = [
    "negligence", "breach of contract", "fraud", "misrepresentation",
    "defamation", "battery", "assault", "false imprisonment",
    "intentional infliction of emotional distress", "conversion",
    "trespass", "nuisance", "unjust enrichment", "breach of warranty",
]


class LegalConceptExtractor:
    """Extract legal concepts from case text."""
    
    def __init__(self):
        self.logger = logger
        
        # Build concept index
        self.concept_index: Dict[str, str] = {}
        for category, concepts in LEGAL_CONCEPTS.items():
            for concept in concepts:
                self.concept_index[concept.lower()] = category
        
        # Add standards
        for category, standards in LEGAL_STANDARDS.items():
            for standard in standards:
                self.concept_index[standard.lower()] = category
        
        # Add causes of action
        for cause in CAUSES_OF_ACTION:
            self.concept_index[cause.lower()] = "causes_of_action"
    
    def extract(self, text: str, min_confidence: float = 0.5) -> List[LegalConcept]:
        """
        Extract legal concepts from text.
        
        Args:
            text: Text to extract from
            min_confidence: Minimum confidence threshold
            
        Returns:
            List of extracted concepts
        """
        text_lower = text.lower()
        extracted = []
        
        for concept, category in self.concept_index.items():
            # Find all occurrences
            pattern = r'\b' + re.escape(concept) + r'\b'
            matches = list(re.finditer(pattern, text_lower))
            
            if matches:
                positions = [m.start() for m in matches]
                
                # Calculate confidence based on frequency and context
                frequency = len(matches)
                confidence = min(frequency / 10.0, 1.0)  # Normalize to 0-1
                
                if confidence >= min_confidence:
                    extracted.append(LegalConcept(
                        concept=concept,
                        category=category,
                        confidence=confidence,
                        positions=positions,
                    ))
        
        # Sort by confidence
        extracted.sort(key=lambda x: x.confidence, reverse=True)
        
        self.logger.info(
            "concepts_extracted",
            count=len(extracted),
            unique_categories=len(set(c.category for c in extracted)),
        )
        
        return extracted
    
    def get_primary_area_of_law(self, text: str) -> Optional[str]:
        """
        Determine the primary area of law from text.
        
        Args:
            text: Text to analyze
            
        Returns:
            Primary area of law or None
        """
        concepts = self.extract(text)
        
        if not concepts:
            return None
        
        # Count concepts by category
        category_scores = Counter()
        for concept in concepts:
            category_scores[concept.category] += concept.confidence
        
        # Return category with highest score
        primary = category_scores.most_common(1)
        return primary[0][0] if primary else None
    
    def get_concept_summary(self, text: str) -> Dict[str, List[str]]:
        """
        Get a summary of concepts by category.
        
        Args:
            text: Text to analyze
            
        Returns:
            Dictionary mapping category to list of concepts
        """
        concepts = self.extract(text)
        
        summary = {}
        for concept in concepts:
            if concept.category not in summary:
                summary[concept.category] = []
            summary[concept.category].append(concept.concept)
        
        return summary
    
    def extract_legal_issues(self, text: str) -> List[str]:
        """
        Extract specific legal issues from text.
        
        Args:
            text: Text to analyze
            
        Returns:
            List of legal issues
        """
        issues = []
        
        # Pattern: "whether..." or "if..."
        issue_patterns = [
            r'whether\s+([^.?!]+)[.?!]',
            r'if\s+([^.?!]+)[.?!]',
            r'the\s+issue\s+is\s+([^.?!]+)[.?!]',
            r'the\s+question\s+is\s+([^.?!]+)[.?!]',
        ]
        
        for pattern in issue_patterns:
            matches = re.finditer(pattern, text, re.IGNORECASE)
            for match in matches:
                issue = match.group(1).strip()
                if len(issue) > 10 and len(issue) < 200:  # Reasonable length
                    issues.append(issue)
        
        return issues[:10]  # Limit to top 10
    
    def tag_case(self, case_text: str) -> Dict[str, any]:
        """
        Generate comprehensive tags for a case.
        
        Args:
            case_text: Full case text
            
        Returns:
            Dictionary of tags and metadata
        """
        concepts = self.extract(case_text)
        
        return {
            "primary_area": self.get_primary_area_of_law(case_text),
            "concepts": [c.concept for c in concepts[:20]],  # Top 20
            "concept_categories": list(set(c.category for c in concepts)),
            "legal_issues": self.extract_legal_issues(case_text),
            "concept_summary": self.get_concept_summary(case_text),
            "total_concepts_found": len(concepts),
        }


def extract_legal_concepts(text: str, min_confidence: float = 0.5) -> List[LegalConcept]:
    """
    Convenience function to extract legal concepts.
    
    Args:
        text: Text to extract from
        min_confidence: Minimum confidence threshold
        
    Returns:
        List of extracted concepts
    """
    extractor = LegalConceptExtractor()
    return extractor.extract(text, min_confidence)


def get_areas_of_law(text: str) -> List[str]:
    """
    Convenience function to get areas of law mentioned in text.
    
    Args:
        text: Text to analyze
        
    Returns:
        List of area names
    """
    extractor = LegalConceptExtractor()
    concepts = extractor.extract(text)
    return list(set(c.category for c in concepts))
