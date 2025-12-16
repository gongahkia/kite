"""Sample data for testing."""
from datetime import datetime
from kite.utils.data_models import CaseData

SAMPLE_CASES = [
    CaseData(
        case_name="Test v. Example",
        case_id="TEST-001",
        court="Test Court",
        date=datetime(2023, 1, 1),
        url="https://example.com/case/1",
        jurisdiction="US",
    ),
]
