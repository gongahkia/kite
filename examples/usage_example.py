from main import extract_document

result = extract_document("examples/sample_contract.pdf", lang="en", translate_to="fr")
print("Jurisdiction:", result['jurisdiction'])
print("Document Type:", result['document_type'])
print("Extracted Data:", result['extracted_data'])
print("Risks:", result['risks'])
print("Text:", result['text'][:200])
print("Metadata:", result['metadata'])
print("Number of Images Extracted:", len(result['images']))