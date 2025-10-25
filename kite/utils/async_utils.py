"""Async utilities for future async support."""
import asyncio
from typing import Callable, Any, List
from concurrent.futures import ThreadPoolExecutor

def run_in_executor(func: Callable, *args: Any, **kwargs: Any) -> Any:
    """Run blocking function in executor."""
    loop = asyncio.get_event_loop()
    with ThreadPoolExecutor() as executor:
        return loop.run_in_executor(executor, lambda: func(*args, **kwargs))

async def gather_with_limit(tasks: List, limit: int = 10):
    """Run tasks with concurrency limit."""
    semaphore = asyncio.Semaphore(limit)
    
    async def bounded_task(task):
        async with semaphore:
            return await task
    
    return await asyncio.gather(*[bounded_task(t) for t in tasks])
