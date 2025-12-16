"""Data serialization utilities."""
from typing import Any, Dict
import json
import pickle
from pathlib import Path

def serialize_to_json(obj: Any, filepath: Path) -> bool:
    """Serialize object to JSON file."""
    try:
        with open(filepath, 'w') as f:
            json.dump(obj, f, default=str, indent=2)
        return True
    except Exception:
        return False

def deserialize_from_json(filepath: Path) -> Any:
    """Deserialize object from JSON file."""
    try:
        with open(filepath, 'r') as f:
            return json.load(f)
    except Exception:
        return None

def serialize_to_pickle(obj: Any, filepath: Path) -> bool:
    """Serialize object to pickle file."""
    try:
        with open(filepath, 'wb') as f:
            pickle.dump(obj, f)
        return True
    except Exception:
        return False

def deserialize_from_pickle(filepath: Path) -> Any:
    """Deserialize object from pickle file."""
    try:
        with open(filepath, 'rb') as f:
            return pickle.load(f)
    except Exception:
        return None
