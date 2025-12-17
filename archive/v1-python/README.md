# Version 1 (Python)

This repository originally shipped Version 1 as a Python library and CLI under `kite/`.

- Language: Python 3
- Entry points: `kite/cli.py`, library modules under `kite/`
- Tests: `tests/`

This folder documents the v1 lineage; source remains in-place under the root for full backward compatibility.

## Usage

### Setup

```console
$ cd archive/v1-python
$ pip install -r requirements.txt
```

### CLI Usage

Extract and analyze a legal document:

```console
$ python src/cli.py path/to/document.pdf
$ python src/cli.py path/to/document.docx --lang en
$ python src/cli.py path/to/document.pdf --translate-to fr
```

### Library Usage

#### Extract Document with Jurisdiction Detection

```python
from src.main import extract_document

result = extract_document(
    "legal_contract.pdf",
    lang="auto"
)

print(f"Jurisdiction: {result['jurisdiction']}")
print(f"Document Type: {result['document_type']}")
print(f"Extracted Data: {result['extracted_data']}")
print(f"Compliance Risks: {result['risks']}")
```

#### Extract and Translate

```python
from src.main import extract_document

result = extract_document(
    "judgment.pdf",
    lang="en",
    translate_to="es"
)

print(f"Translated Text: {result['text']}")
```

#### PDF Extraction

```python
from src.extractors.pdf_extractor import PDFExtractor

extractor = PDFExtractor("case_file.pdf")
text, images, metadata = extractor.extract_all()
jurisdiction = extractor.detect_jurisdiction(metadata, text)

print(f"Extracted {len(images)} images")
print(f"Detected jurisdiction: {jurisdiction}")
```

#### DOCX Extraction

```python
from src.extractors.docx_extractor import DOCXExtractor

extractor = DOCXExtractor("legal_memo.docx")
text, images, metadata = extractor.extract_all()

print(f"Text length: {len(text)} characters")
print(f"Metadata: {metadata}")
```

#### Document Classification

```python
from src.nlp.classifier import DocumentClassifier

text = "This is a legal contract between..."
doc_type = DocumentClassifier.classify(text, lang="en")

print(f"Document Type: {doc_type}")
```

#### Compliance Checking

```python
from src.nlp.compliance import ComplianceChecker
from src.config.jurisdictions import get_jurisdiction_config

config = get_jurisdiction_config("US")
risks = ComplianceChecker.check(text, config, doc_type="contract")

for risk in risks:
    print(f"⚠️  {risk['severity']}: {risk['description']}")
```

### REST API Usage

Version 1 includes a Flask-based API:

```python
from src.api import run

run()  # Starts server on http://0.0.0.0:5000
```

Then make requests:

```console
$ curl -X POST http://localhost:5000/extract \
  -F "file=@document.pdf" \
  -F "lang=en" \
  -F "translate_to=fr"
```

### Jurisdiction Configuration

```python
from src.config.jurisdictions import get_jurisdiction_config

us_config = get_jurisdiction_config("US")
print(f"Rules: {us_config['rules']}")
print(f"Date formats: {us_config['date_formats']}")
```
