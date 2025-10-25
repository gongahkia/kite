"""Mock helpers for testing."""
from unittest.mock import Mock
import requests

def create_mock_response(status_code: int = 200, text: str = ""):
    """Create mock HTTP response."""
    response = Mock(spec=requests.Response)
    response.status_code = status_code
    response.text = text
    response.headers = {}
    return response
