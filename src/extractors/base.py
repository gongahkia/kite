class BaseExtractor:
    def __init__(self, file_path):
        self.file_path = file_path
        self.text = None
        self.images = []
        self.metadata = {}

    def extract_all(self):
        raise NotImplementedError("extract_all must be implemented by subclasses")

    def detect_jurisdiction(self, metadata, text):
        if 'US' in text or 'United States' in text:
            return 'US'
        elif 'EU' in text or 'European Union' in text:
            return 'EU'
        elif 'UK' in text or 'United Kingdom' in text:
            return 'UK'
        elif 'BR' in text or 'Brazil' in text:
            return 'BR'
        elif 'CN' in text or 'China' in text:
            return 'CN'
        elif 'DE' in text or 'Germany' in text:
            return 'DE'
        elif 'FR' in text or 'France' in text:
            return 'FR'
        elif 'JP' in text or 'Japan' in text:
            return 'JP'
        elif 'MS' in text or 'Malaysia' in text:
            return 'MS'
        elif 'SG' in text or 'Singapore' in text:
            return 'SG'
        else:
            return None

    def apply_rules(self, config, doc_type):
        rules = config.get('document_types', {}).get(doc_type, {})
        required_fields = rules.get('required_fields', [])
        extracted_data = {field: None for field in required_fields}
        return extracted_data