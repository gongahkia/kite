"""URL validation utilities."""
from urllib.parse import urlparse
from typing import Optional

def is_valid_url(url: str) -> bool:
    """Check if URL is valid."""
    try:
        result = urlparse(url)
        return all([result.scheme, result.netloc])
    except Exception:
        return False

def normalize_url(url: str) -> Optional[str]:
    """Normalize URL to standard format."""
    if not is_valid_url(url):
        return None
    
    parsed = urlparse(url)
    
    # Ensure https
    if parsed.scheme == "http":
        parsed = parsed._replace(scheme="https")
    
    # Remove trailing slash
    path = parsed.path.rstrip("/") if parsed.path != "/" else "/"
    parsed = parsed._replace(path=path)
    
    return parsed.geturl()
