from docx import Document
from PIL import Image
import io
from .base import BaseExtractor

class DOCXExtractor(BaseExtractor):
    def __init__(self, file_path):
        super().__init__(file_path)
        self.doc = Document(file_path)

    def extract_all(self):
        text = []
        images = []
        metadata = {}

        for para in self.doc.paragraphs:
            text.append(para.text)

        for rel in self.doc.part._rels:
            rel = self.doc.part._rels[rel]
            if "image" in rel.target_ref:
                image_data = rel.target_part.blob
                image = Image.open(io.BytesIO(image_data))
                images.append(image)

        self.text = "\n".join(text)
        self.images = images
        self.metadata = metadata
        return self.text, self.images, self.metadata

    def detect_jurisdiction(self, metadata, text):
        if 'United States' in text or 'US' in text:
            return 'US'
        elif 'European Union' in text or 'EU' in text:
            return 'EU'
        elif 'United Kingdom' in text or 'UK' in text:
            return 'UK'
        elif 'Brazil' in text or 'BR' in text:
            return 'BR'
        elif 'China' in text or 'CN' in text:
            return 'CN'
        elif 'Germany' in text or 'DE' in text:
            return 'DE'
        elif 'France' in text or 'FR' in text:
            return 'FR'
        elif 'Japan' in text or 'JP' in text:
            return 'JP'
        elif 'Malaysia' in text or 'MS' in text:
            return 'MS'
        elif 'Singapore' in text or 'SG' in text:
            return 'SG'
        return super().detect_jurisdiction(metadata, text)

    def apply_rules(self, config, doc_type):
        return super().apply_rules(config, doc_type)