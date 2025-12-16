#!/usr/bin/env python3
"""
Verify package setup before publishing to PyPI.
"""

import sys
import subprocess
from pathlib import Path


def run_command(cmd, description):
    """Run a command and report results."""
    print(f"\n{'='*60}")
    print(f"Checking: {description}")
    print(f"{'='*60}")
    
    result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
    
    if result.returncode == 0:
        print(f"✓ {description} - PASSED")
        if result.stdout:
            print(result.stdout)
        return True
    else:
        print(f"✗ {description} - FAILED")
        if result.stderr:
            print(result.stderr)
        return False


def check_files_exist():
    """Check that required files exist."""
    print(f"\n{'='*60}")
    print("Checking: Required files exist")
    print(f"{'='*60}")
    
    required_files = [
        'README.md',
        'LICENSE',
        'pyproject.toml',
        'requirements.txt',
        'MANIFEST.in',
        'kite/__init__.py',
    ]
    
    all_exist = True
    for file_path in required_files:
        path = Path(file_path)
        if path.exists():
            print(f"✓ {file_path}")
        else:
            print(f"✗ {file_path} - MISSING")
            all_exist = False
    
    if all_exist:
        print("\n✓ All required files exist - PASSED")
    else:
        print("\n✗ Some required files are missing - FAILED")
    
    return all_exist


def main():
    """Run all verification checks."""
    print("Kite Package Setup Verification")
    print("="*60)
    
    checks = []
    
    # Check required files
    checks.append(check_files_exist())
    
    # Check if build tools are installed
    checks.append(run_command(
        "python -m pip show build",
        "Build tools installed"
    ))
    
    # Check if twine is installed
    checks.append(run_command(
        "python -m pip show twine",
        "Twine installed"
    ))
    
    # Try to build the package
    print(f"\n{'='*60}")
    print("Building package (dry run)...")
    print(f"{'='*60}")
    
    # Clean previous builds
    subprocess.run("rm -rf dist/ build/ *.egg-info", shell=True)
    
    checks.append(run_command(
        "python -m build",
        "Package build"
    ))
    
    # Check package with twine
    if Path("dist").exists():
        checks.append(run_command(
            "python -m twine check dist/*",
            "Package validation with twine"
        ))
    
    # Final summary
    print(f"\n{'='*60}")
    print("VERIFICATION SUMMARY")
    print(f"{'='*60}")
    
    passed = sum(checks)
    total = len(checks)
    
    print(f"Checks passed: {passed}/{total}")
    
    if passed == total:
        print("\n✓ ALL CHECKS PASSED - Ready to publish!")
        print("\nNext steps:")
        print("1. Test on TestPyPI: python -m twine upload --repository testpypi dist/*")
        print("2. If successful, publish to PyPI: python -m twine upload dist/*")
        return 0
    else:
        print("\n✗ SOME CHECKS FAILED - Fix issues before publishing")
        return 1


if __name__ == "__main__":
    sys.exit(main())
