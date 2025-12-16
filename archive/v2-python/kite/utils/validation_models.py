"""
Pydantic models for data validation and quality assurance.
"""

from datetime import datetime
from typing import Optional, List, Dict, Any
from enum import Enum
from pydantic import BaseModel, Field, field_validator, model_validator, HttpUrl
from pydantic import ConfigDict


class CourtLevel(str, Enum):
    """Enumeration of court hierarchy levels."""
    SUPREME = "supreme"
    APPELLATE = "appellate"
    TRIAL = "trial"
    DISTRICT = "district"
    MAGISTRATE = "magistrate"
    TRIBUNAL = "tribunal"
    INTERNATIONAL = "international"
    OTHER = "other"


class CaseOutcome(str, Enum):
    """Enumeration of case outcomes."""
    AFFIRMED = "affirmed"
    REVERSED = "reversed"
    REMANDED = "remanded"
    DISMISSED = "dismissed"
    GRANTED = "granted"
    DENIED = "denied"
    SETTLED = "settled"
    WITHDRAWN = "withdrawn"
    PENDING = "pending"
    OTHER = "other"


class PrecedentialValue(str, Enum):
    """Precedential value of a case."""
    BINDING = "binding"
    PERSUASIVE = "persuasive"
    NON_PRECEDENTIAL = "non-precedential"
    UNPUBLISHED = "unpublished"
    UNKNOWN = "unknown"


class ValidatedCaseData(BaseModel):
    """
    Validated case data model using Pydantic.
    Provides automatic validation, type checking, and serialization.
    """
    
    model_config = ConfigDict(
        str_strip_whitespace=True,
        validate_assignment=True,
        use_enum_values=True,
    )
    
    # Core fields (required)
    case_name: str = Field(..., min_length=1, max_length=1000, description="Name of the case")
    jurisdiction: str = Field(..., min_length=1, description="Legal jurisdiction")
    
    # Identifiers
    case_id: Optional[str] = Field(None, description="Unique case identifier")
    docket_number: Optional[str] = Field(None, description="Court docket number")
    
    # Court information
    court: Optional[str] = Field(None, max_length=500, description="Name of the court")
    court_level: Optional[CourtLevel] = Field(None, description="Level in court hierarchy")
    
    # Date information
    date: Optional[datetime] = Field(None, description="Decision date")
    filing_date: Optional[datetime] = Field(None, description="Case filing date")
    
    # URLs and sources
    url: Optional[HttpUrl] = Field(None, description="URL to the case")
    pdf_url: Optional[HttpUrl] = Field(None, description="URL to PDF version")
    
    # Case content
    full_text: Optional[str] = Field(None, description="Full text of the decision")
    summary: Optional[str] = Field(None, max_length=5000, description="Case summary")
    holding: Optional[str] = Field(None, description="Main holding/ratio decidendi")
    
    # Parties and judges
    judges: List[str] = Field(default_factory=list, description="Judges on the case")
    parties: List[str] = Field(default_factory=list, description="Parties involved")
    counsel: List[str] = Field(default_factory=list, description="Counsel/attorneys")
    
    # Legal metadata
    legal_issues: List[str] = Field(default_factory=list, description="Legal issues addressed")
    citations: List[str] = Field(default_factory=list, description="Citations to this case")
    cited_cases: List[str] = Field(default_factory=list, description="Cases cited by this case")
    statutes_cited: List[str] = Field(default_factory=list, description="Statutes/codes cited")
    
    # Case categorization
    case_type: Optional[str] = Field(None, description="Type of case (civil, criminal, etc.)")
    areas_of_law: List[str] = Field(default_factory=list, description="Areas of law")
    outcome: Optional[CaseOutcome] = Field(None, description="Case outcome/disposition")
    precedential_value: Optional[PrecedentialValue] = Field(
        PrecedentialValue.UNKNOWN,
        description="Precedential value"
    )
    
    # Additional metadata
    panel_size: Optional[int] = Field(None, ge=1, le=50, description="Number of judges on panel")
    is_published: bool = Field(True, description="Whether case is officially published")
    language: str = Field("en", description="Language of the decision")
    
    # Data quality
    scraped_at: datetime = Field(default_factory=datetime.now, description="When data was scraped")
    completeness_score: Optional[float] = Field(None, ge=0.0, le=1.0, description="Data completeness (0-1)")
    
    # Flexible metadata
    metadata: Dict[str, Any] = Field(default_factory=dict, description="Additional metadata")
    
    @field_validator('case_name')
    @classmethod
    def validate_case_name(cls, v: str) -> str:
        """Validate and clean case name."""
        if not v or not v.strip():
            raise ValueError("Case name cannot be empty")
        
        # Remove excessive whitespace
        v = ' '.join(v.split())
        
        return v
    
    @field_validator('judges', 'parties', 'legal_issues', 'citations')
    @classmethod
    def remove_empty_strings(cls, v: List[str]) -> List[str]:
        """Remove empty strings from lists."""
        return [item.strip() for item in v if item and item.strip()]
    
    @model_validator(mode='after')
    def validate_dates(self):
        """Ensure date consistency."""
        if self.filing_date and self.date:
            if self.filing_date > self.date:
                raise ValueError("Filing date cannot be after decision date")
        
        if self.date and self.date > datetime.now():
            raise ValueError("Decision date cannot be in the future")
        
        return self
    
    @model_validator(mode='after')
    def calculate_completeness(self):
        """Calculate data completeness score."""
        if self.completeness_score is None:
            total_fields = 0
            filled_fields = 0
            
            # Core fields
            for field_name in ['case_name', 'jurisdiction', 'case_id', 'court', 
                              'date', 'url', 'summary']:
                total_fields += 1
                if getattr(self, field_name):
                    filled_fields += 1
            
            # List fields
            for field_name in ['judges', 'parties', 'legal_issues', 'citations']:
                total_fields += 1
                if getattr(self, field_name):
                    filled_fields += 1
            
            self.completeness_score = filled_fields / total_fields if total_fields > 0 else 0.0
        
        return self
    
    def to_legacy_format(self) -> Dict[str, Any]:
        """Convert to legacy CaseData format for backwards compatibility."""
        from .data_models import CaseData
        
        return CaseData(
            case_name=self.case_name,
            case_id=self.case_id,
            court=self.court,
            date=self.date,
            url=str(self.url) if self.url else None,
            full_text=self.full_text,
            summary=self.summary,
            judges=self.judges,
            parties=self.parties,
            legal_issues=self.legal_issues,
            citations=self.citations,
            jurisdiction=self.jurisdiction,
            case_type=self.case_type,
            outcome=self.outcome.value if self.outcome else None,
            metadata=self.metadata,
        )


class ValidationResult(BaseModel):
    """Result of data validation."""
    
    is_valid: bool = Field(..., description="Whether data passed validation")
    errors: List[str] = Field(default_factory=list, description="Validation errors")
    warnings: List[str] = Field(default_factory=list, description="Validation warnings")
    completeness_score: float = Field(..., ge=0.0, le=1.0, description="Completeness score")
    validated_data: Optional[ValidatedCaseData] = Field(None, description="Validated data if successful")
    
    def __bool__(self) -> bool:
        """Allow boolean evaluation."""
        return self.is_valid


class DataQualityMetrics(BaseModel):
    """Metrics for data quality assessment."""
    
    total_cases: int = Field(0, description="Total number of cases")
    valid_cases: int = Field(0, description="Number of valid cases")
    invalid_cases: int = Field(0, description="Number of invalid cases")
    
    avg_completeness: float = Field(0.0, ge=0.0, le=1.0, description="Average completeness score")
    fields_coverage: Dict[str, float] = Field(
        default_factory=dict,
        description="Coverage percentage by field"
    )
    
    validation_errors: Dict[str, int] = Field(
        default_factory=dict,
        description="Count of each type of validation error"
    )
    
    @property
    def validation_rate(self) -> float:
        """Calculate validation pass rate."""
        if self.total_cases == 0:
            return 0.0
        return self.valid_cases / self.total_cases
    
    @property
    def quality_grade(self) -> str:
        """Assign a quality grade based on metrics."""
        rate = self.validation_rate
        if rate >= 0.95:
            return "A"
        elif rate >= 0.90:
            return "B"
        elif rate >= 0.80:
            return "C"
        elif rate >= 0.70:
            return "D"
        else:
            return "F"
