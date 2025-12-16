"""Text processing utilities."""
import re
from typing import List, Optional

def clean_whitespace(text: str) -> str:
    """Remove extra whitespace from text."""
    return " ".join(text.split())

def remove_html_tags(text: str) -> str:
    """Remove HTML tags from text."""
    clean = re.compile('<.*?>')
    return re.sub(clean, '', text)

def truncate_text(text: str, max_length: int = 200, suffix: str = "...") -> str:
    """Truncate text to max length with suffix."""
    if len(text) <= max_length:
        return text
    return text[:max_length - len(suffix)] + suffix

def extract_emails(text: str) -> List[str]:
    """Extract email addresses from text."""
    pattern = r'\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b'
    return re.findall(pattern, text)

def sanitize_filename(filename: str) -> str:
    """Sanitize filename by removing invalid characters."""
    return re.sub(r'[^\w\s.-]', '', filename)
