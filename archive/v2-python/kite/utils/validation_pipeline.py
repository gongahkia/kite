"""
Data validation pipeline for ensuring data quality.
"""

from typing import List, Dict, Any, Optional
from datetime import datetime
from pydantic import ValidationError

from .validation_models import (
    ValidatedCaseData,
    ValidationResult,
    DataQualityMetrics,
    CourtLevel,
    CaseOutcome,
    PrecedentialValue,
)
from .data_models import CaseData
from .logging_config import get_logger

logger = get_logger(__name__)


class ValidationPipeline:
    """
    Pipeline for validating case data with multiple validation stages.
    """
    
    def __init__(self, strict: bool = False):
        """
        Initialize validation pipeline.
        
        Args:
            strict: If True, fail on warnings; if False, warnings don't fail validation
        """
        self.strict = strict
        self.logger = logger
    
    def validate_case(
        self,
        case_data: Dict[str, Any],
        source: str = "unknown"
    ) -> ValidationResult:
        """
        Validate a single case.
        
        Args:
            case_data: Dictionary containing case data
            source: Source of the data (for logging)
            
        Returns:
            ValidationResult with validation status and any errors
        """
        errors = []
        warnings = []
        
        try:
            # Try to create validated model
            validated = ValidatedCaseData(**case_data)
            
            # Run business rule validations
            biz_warnings = self._validate_business_rules(validated)
            warnings.extend(biz_warnings)
            
            # Check for suspicious data patterns
            suspicious_warnings = self._check_suspicious_patterns(validated)
            warnings.extend(suspicious_warnings)
            
            is_valid = True if not self.strict else (len(warnings) == 0)
            
            return ValidationResult(
                is_valid=is_valid,
                errors=errors,
                warnings=warnings,
                completeness_score=validated.completeness_score or 0.0,
                validated_data=validated,
            )
            
        except ValidationError as e:
            # Collect all validation errors
            for error in e.errors():
                field = ".".join(str(x) for x in error["loc"])
                msg = error["msg"]
                errors.append(f"{field}: {msg}")
            
            self.logger.warning(
                "validation_failed",
                source=source,
                errors=errors,
            )
            
            return ValidationResult(
                is_valid=False,
                errors=errors,
                warnings=warnings,
                completeness_score=0.0,
            )
    
    def _validate_business_rules(self, case: ValidatedCaseData) -> List[str]:
        """
        Validate business rules specific to legal data.
        
        Args:
            case: Validated case data
            
        Returns:
            List of warning messages
        """
        warnings = []
        
        # Check for very old cases without proper metadata
        if case.date and case.date.year < 1800:
            warnings.append(f"Case date is unusually old: {case.date.year}")
        
        # Check for unreasonably long case names
        if len(case.case_name) > 500:
            warnings.append("Case name is unusually long (>500 chars)")
        
        # Check for missing critical fields
        if not case.case_id and not case.docket_number:
            warnings.append("No case ID or docket number provided")
        
        if not case.court:
            warnings.append("Court name is missing")
        
        if not case.date:
            warnings.append("Decision date is missing")
        
        # Check for empty lists that should have data
        if case.is_published and not case.citations:
            warnings.append("Published case has no citations")
        
        # Check citation format (basic check)
        for citation in case.citations:
            if len(citation) < 5 or not any(c.isdigit() for c in citation):
                warnings.append(f"Suspicious citation format: {citation}")
        
        return warnings
    
    def _check_suspicious_patterns(self, case: ValidatedCaseData) -> List[str]:
        """
        Check for suspicious data patterns that might indicate scraping issues.
        
        Args:
            case: Validated case data
            
        Returns:
            List of warning messages
        """
        warnings = []
        
        # Check for HTML tags in text (scraping artifact)
        for field_name in ['case_name', 'summary', 'holding']:
            value = getattr(case, field_name, None)
            if value and isinstance(value, str):
                if '<' in value and '>' in value:
                    warnings.append(f"{field_name} contains HTML tags")
        
        # Check for repeated text (pagination artifacts)
        if case.full_text:
            words = case.full_text.split()
            if len(words) > 50:
                # Check for high repetition
                unique_words = len(set(words))
                if unique_words < len(words) * 0.3:
                    warnings.append("Full text has unusually high word repetition")
        
        # Check for placeholder text
        placeholder_texts = [
            'lorem ipsum', 'test', 'example', '[placeholder]',
            'todo', 'tbd', 'n/a', 'null', 'none', 'unknown'
        ]
        
        for field_name in ['case_name', 'summary', 'court']:
            value = getattr(case, field_name, None)
            if value and isinstance(value, str):
                value_lower = value.lower()
                for placeholder in placeholder_texts:
                    if placeholder in value_lower:
                        warnings.append(f"{field_name} contains placeholder text: '{placeholder}'")
        
        return warnings
    
    def validate_batch(
        self,
        cases: List[Dict[str, Any]],
        source: str = "unknown"
    ) -> DataQualityMetrics:
        """
        Validate a batch of cases and generate quality metrics.
        
        Args:
            cases: List of case data dictionaries
            source: Source of the data
            
        Returns:
            DataQualityMetrics with aggregated statistics
        """
        metrics = DataQualityMetrics(total_cases=len(cases))
        
        completeness_scores = []
        field_counts: Dict[str, int] = {}
        error_counts: Dict[str, int] = {}
        
        for idx, case_data in enumerate(cases):
            result = self.validate_case(case_data, source=f"{source}[{idx}]")
            
            if result.is_valid:
                metrics.valid_cases += 1
                if result.validated_data:
                    completeness_scores.append(result.validated_data.completeness_score or 0.0)
                    
                    # Track field coverage
                    for field_name, value in result.validated_data.model_dump().items():
                        if value is not None and value != [] and value != {}:
                            field_counts[field_name] = field_counts.get(field_name, 0) + 1
            else:
                metrics.invalid_cases += 1
                
                # Track error types
                for error in result.errors:
                    error_type = error.split(':')[0] if ':' in error else error
                    error_counts[error_type] = error_counts.get(error_type, 0) + 1
        
        # Calculate averages
        if completeness_scores:
            metrics.avg_completeness = sum(completeness_scores) / len(completeness_scores)
        
        # Calculate field coverage percentages
        for field_name, count in field_counts.items():
            metrics.fields_coverage[field_name] = (count / len(cases)) * 100
        
        metrics.validation_errors = error_counts
        
        self.logger.info(
            "batch_validation_complete",
            source=source,
            total=metrics.total_cases,
            valid=metrics.valid_cases,
            invalid=metrics.invalid_cases,
            avg_completeness=metrics.avg_completeness,
            quality_grade=metrics.quality_grade,
        )
        
        return metrics
    
    def convert_legacy_case(self, case: CaseData) -> ValidationResult:
        """
        Convert legacy CaseData to validated format.
        
        Args:
            case: Legacy CaseData object
            
        Returns:
            ValidationResult with converted data
        """
        case_dict = {
            'case_name': case.case_name,
            'jurisdiction': case.jurisdiction or 'unknown',
            'case_id': case.case_id,
            'court': case.court,
            'date': case.date,
            'url': case.url,
            'full_text': case.full_text,
            'summary': case.summary,
            'judges': case.judges or [],
            'parties': case.parties or [],
            'legal_issues': case.legal_issues or [],
            'citations': case.citations or [],
            'case_type': case.case_type,
            'outcome': case.outcome,
            'metadata': case.metadata or {},
        }
        
        return self.validate_case(case_dict, source="legacy_conversion")


def validate_case_data(case_data: Dict[str, Any], strict: bool = False) -> ValidationResult:
    """
    Convenience function to validate a single case.
    
    Args:
        case_data: Dictionary containing case data
        strict: If True, warnings cause validation failure
        
    Returns:
        ValidationResult
    """
    pipeline = ValidationPipeline(strict=strict)
    return pipeline.validate_case(case_data)


def validate_case_batch(cases: List[Dict[str, Any]], strict: bool = False) -> DataQualityMetrics:
    """
    Convenience function to validate a batch of cases.
    
    Args:
        cases: List of case data dictionaries
        strict: If True, warnings cause validation failure
        
    Returns:
        DataQualityMetrics
    """
    pipeline = ValidationPipeline(strict=strict)
    return pipeline.validate_batch(cases)
