from extractors.pdf_extractor import PDFExtractor
from extractors.docx_extractor import DOCXExtractor
from nlp.classifier import DocumentClassifier
from nlp.compliance import ComplianceChecker
from nlp.translation import Translator
from config.jurisdictions import get_jurisdiction_config

def extract_document(file_path, lang='auto', translate_to=None):
    if file_path.endswith('.pdf'):
        extractor = PDFExtractor(file_path)
    elif file_path.endswith('.docx'):
        extractor = DOCXExtractor(file_path)
    else:
        raise ValueError("Unsupported file format")
    
    text, images, metadata = extractor.extract_all()
    jurisdiction = extractor.detect_jurisdiction(metadata, text)
    config = get_jurisdiction_config(jurisdiction)
    doc_type = DocumentClassifier.classify(text, lang=lang)
    extracted_data = extractor.apply_rules(config, doc_type)
    risks = ComplianceChecker.check(text, config, doc_type)
    if translate_to:
        text = Translator.translate(text, target_lang=translate_to)
    return {
        'jurisdiction': jurisdiction,
        'document_type': doc_type,
        'extracted_data': extracted_data,
        'images': images,
        'risks': risks,
        'text': text,
        'metadata': metadata
    }