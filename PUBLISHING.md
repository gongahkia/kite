# Publishing to PyPI

This guide explains how to publish Kite to PyPI.

## Prerequisites

1. Create accounts:
   - [PyPI](https://pypi.org/account/register/)
   - [TestPyPI](https://test.pypi.org/account/register/) (for testing)

2. Install build tools:
```bash
pip install build twine
```

3. Set up API tokens:
   - Go to PyPI Account Settings > API tokens
   - Create a token for the `kite` project
   - Save it securely

## Building the Package

1. Clean previous builds:
```bash
rm -rf dist/ build/ *.egg-info
```

2. Build the package:
```bash
python -m build
```

This creates both wheel (.whl) and source (.tar.gz) distributions in the `dist/` directory.

## Testing on TestPyPI

1. Upload to TestPyPI first:
```bash
python -m twine upload --repository testpypi dist/*
```

2. Test installation:
```bash
pip install --index-url https://test.pypi.org/simple/ --extra-index-url https://pypi.org/simple/ kite
```

3. Verify the package works:
```bash
python -c "import kite; print(kite.__version__)"
```

## Publishing to PyPI

1. Upload to PyPI:
```bash
python -m twine upload dist/*
```

2. Enter your API token when prompted (use `__token__` as username)

3. Verify on PyPI: https://pypi.org/project/kite/

4. Test installation:
```bash
pip install kite
```

## Automated Publishing with GitHub Actions

Create `.github/workflows/publish.yml` (already in repo) to automate publishing on release tags.

### Creating a Release

1. Update version in `pyproject.toml` and `kite/__init__.py`

2. Commit and tag:
```bash
git add pyproject.toml kite/__init__.py
git commit -m "Bump version to X.Y.Z"
git tag vX.Y.Z
git push origin main --tags
```

3. GitHub Actions will automatically build and publish to PyPI

## Troubleshooting

### Version Already Exists
PyPI doesn't allow re-uploading the same version. Bump the version number and try again.

### Import Errors
Ensure all dependencies are listed in `pyproject.toml` under `dependencies`.

### Missing Files
Check `MANIFEST.in` includes all necessary files.

## Version Numbering

Follow [Semantic Versioning](https://semver.org/):
- MAJOR version for incompatible API changes
- MINOR version for backwards-compatible functionality
- PATCH version for backwards-compatible bug fixes

Current version: 1.0.0
