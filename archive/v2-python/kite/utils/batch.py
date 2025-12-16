"""Batch processing utilities."""
from typing import List, Callable, TypeVar, Iterable

T = TypeVar('T')
R = TypeVar('R')

def batch_process(
    items: List[T],
    batch_size: int,
    processor: Callable[[List[T]], List[R]]
) -> List[R]:
    """Process items in batches."""
    results = []
    
    for i in range(0, len(items), batch_size):
        batch = items[i:i + batch_size]
        batch_results = processor(batch)
        results.extend(batch_results)
    
    return results

def chunked(iterable: Iterable[T], size: int) -> Iterable[List[T]]:
    """Split iterable into chunks."""
    chunk = []
    
    for item in iterable:
        chunk.append(item)
        if len(chunk) >= size:
            yield chunk
            chunk = []
    
    if chunk:
        yield chunk
