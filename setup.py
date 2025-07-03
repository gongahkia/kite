from setuptools import setup, find_packages

setup(
    name="kite",
    version="0.1.0",
    description="KITE: Jurisdiction-aware legal document image extraction library",
    author="Your Name",
    author_email="your.email@example.com",
    package_dir={"": "src"},
    packages=find_packages(where="src"),
    include_package_data=True,
    install_requires=[
        "flask",
        "python-docx",
        "PyMuPDF",
        "pytesseract",
        "Pillow",
        "spacy",
        "googletrans==4.0.0rc1",
        "pyyaml",
    ],
    entry_points={
        "console_scripts": [
            "kite=src.cli:main",
        ],
    },
    python_requires=">=3.8",
    classifiers=[
        "Programming Language :: Python :: 3",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
    ],
)