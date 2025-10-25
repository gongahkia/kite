"""String manipulation utilities."""
import unicodedata
from typing import Optional

def slugify(text: str) -> str:
    """Convert text to URL-safe slug."""
    text = unicodedata.normalize('NFKD', text)
    text = text.encode('ascii', 'ignore').decode('ascii')
    text = text.lower()
    text = ''.join(c if c.isalnum() or c in '-_' else '-' for c in text)
    return text.strip('-')

def camel_to_snake(text: str) -> str:
    """Convert camelCase to snake_case."""
    import re
    text = re.sub('(.)([A-Z][a-z]+)', r'\1_\2', text)
    return re.sub('([a-z0-9])([A-Z])', r'\1_\2', text).lower()

def snake_to_camel(text: str) -> str:
    """Convert snake_case to camelCase."""
    components = text.split('_')
    return components[0] + ''.join(x.title() for x in components[1:])

def title_case(text: str) -> str:
    """Convert text to title case."""
    return ' '.join(word.capitalize() for word in text.split())
