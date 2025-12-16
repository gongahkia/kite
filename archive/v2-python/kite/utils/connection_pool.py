"""Connection pooling for HTTP requests."""
import requests
from requests.adapters import HTTPAdapter
from urllib3.util.retry import Retry

def create_session(
    max_retries: int = 3,
    backoff_factor: float = 0.3,
    pool_connections: int = 10,
    pool_maxsize: int = 10
) -> requests.Session:
    """Create a requests session with connection pooling and retries."""
    session = requests.Session()
    
    retry_strategy = Retry(
        total=max_retries,
        backoff_factor=backoff_factor,
        status_forcelist=[429, 500, 502, 503, 504],
        method_whitelist=["HEAD", "GET", "OPTIONS"]
    )
    
    adapter = HTTPAdapter(
        max_retries=retry_strategy,
        pool_connections=pool_connections,
        pool_maxsize=pool_maxsize
    )
    
    session.mount("http://", adapter)
    session.mount("https://", adapter)
    
    return session
