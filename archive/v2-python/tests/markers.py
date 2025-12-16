"""Pytest markers."""
import pytest

# Mark slow tests
slow = pytest.mark.slow

# Mark integration tests
integration = pytest.mark.integration

# Mark unit tests  
unit = pytest.mark.unit

# Mark tests that require network
requires_network = pytest.mark.requires_network

# Mark tests that require auth
requires_auth = pytest.mark.requires_auth
