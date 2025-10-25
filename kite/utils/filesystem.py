"""File system utilities."""
from pathlib import Path
from typing import List, Optional
import shutil

def ensure_directory(path: Path) -> bool:
    """Ensure directory exists."""
    try:
        path.mkdir(parents=True, exist_ok=True)
        return True
    except Exception:
        return False

def list_files(directory: Path, pattern: str = "*") -> List[Path]:
    """List files in directory matching pattern."""
    if not directory.is_dir():
        return []
    return list(directory.glob(pattern))

def safe_delete(path: Path) -> bool:
    """Safely delete file or directory."""
    try:
        if path.is_file():
            path.unlink()
        elif path.is_dir():
            shutil.rmtree(path)
        return True
    except Exception:
        return False

def get_file_size(path: Path) -> Optional[int]:
    """Get file size in bytes."""
    try:
        return path.stat().st_size if path.is_file() else None
    except Exception:
        return None
