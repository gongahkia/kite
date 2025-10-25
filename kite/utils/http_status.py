"""HTTP status code utilities."""
from typing import Optional

def is_success(status_code: int) -> bool:
    """Check if status code is success (2xx)."""
    return 200 <= status_code < 300

def is_redirect(status_code: int) -> bool:
    """Check if status code is redirect (3xx)."""
    return 300 <= status_code < 400

def is_client_error(status_code: int) -> bool:
    """Check if status code is client error (4xx)."""
    return 400 <= status_code < 500

def is_server_error(status_code: int) -> bool:
    """Check if status code is server error (5xx)."""
    return 500 <= status_code < 600

def get_status_message(status_code: int) -> Optional[str]:
    """Get status message for code."""
    STATUS_MESSAGES = {
        200: "OK",
        201: "Created",
        204: "No Content",
        400: "Bad Request",
        401: "Unauthorized",
        403: "Forbidden",
        404: "Not Found",
        429: "Too Many Requests",
        500: "Internal Server Error",
        502: "Bad Gateway",
        503: "Service Unavailable",
    }
    return STATUS_MESSAGES.get(status_code)
