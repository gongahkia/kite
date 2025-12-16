"""JSON utilities."""
import json
from typing import Any, Optional
from pathlib import Path

def safe_json_loads(data: str) -> Optional[dict]:
    """Safely load JSON with error handling."""
    try:
        return json.loads(data)
    except json.JSONDecodeError:
        return None

def pretty_print_json(data: dict, indent: int = 2) -> str:
    """Pretty print JSON data."""
    return json.dumps(data, indent=indent, ensure_ascii=False, sort_keys=True)

def save_json(data: dict, filepath: Path) -> bool:
    """Save data to JSON file."""
    try:
        with open(filepath, 'w', encoding='utf-8') as f:
            json.dump(data, f, indent=2, ensure_ascii=False)
        return True
    except Exception:
        return False

def load_json(filepath: Path) -> Optional[dict]:
    """Load data from JSON file."""
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            return json.load(f)
    except Exception:
        return None
