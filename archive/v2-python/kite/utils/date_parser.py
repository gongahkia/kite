"""Date parsing utilities."""
from datetime import datetime
from typing import Optional
import dateutil.parser

def parse_date(date_str: str) -> Optional[datetime]:
    """Parse date string to datetime object."""
    try:
        return dateutil.parser.parse(date_str)
    except (ValueError, TypeError):
        return None

def format_date(date: datetime, format_str: str = "%Y-%m-%d") -> str:
    """Format datetime object to string."""
    return date.strftime(format_str)

def is_valid_date_range(start: str, end: str) -> bool:
    """Check if date range is valid."""
    start_date = parse_date(start)
    end_date = parse_date(end)
    
    if not start_date or not end_date:
        return False
    
    return start_date <= end_date
