"""CLI helper functions."""
import sys
from typing import Optional

def print_error(message: str) -> None:
    """Print error message to stderr."""
    print(f"Error: {message}", file=sys.stderr)

def print_success(message: str) -> None:
    """Print success message."""
    print(f"✓ {message}")

def print_warning(message: str) -> None:
    """Print warning message."""
    print(f"⚠ {message}")

def confirm(message: str, default: bool = False) -> bool:
    """Prompt user for confirmation."""
    suffix = " [Y/n]" if default else " [y/N]"
    response = input(message + suffix + ": ").lower()
    
    if not response:
        return default
    
    return response in ('y', 'yes')
