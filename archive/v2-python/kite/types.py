"""Type definitions."""
from typing import Dict, List, Optional, Union, Any
from datetime import datetime

# Type aliases
SearchQuery = str
CaseID = str
CourtName = str
Jurisdiction = str
URL = str

# Common types
Headers = Dict[str, str]
Params = Dict[str, Union[str, int, float]]
JSONData = Dict[str, Any]

# Result types
SearchResults = List[Dict[str, Any]]
CaseDetails = Dict[str, Any]

# Time types
DateString = str
DateRange = tuple[Optional[datetime], Optional[datetime]]
