import fitz
import pytesseract
from PIL import Image
import io
from .base import BaseExtractor

class PDFExtractor(BaseExtractor):
    def __init__(self, file_path):
        super().__init__(file_path)
        self.doc = fitz.open(file_path)

    def extract_all(self):
        text = []
        images = []
        metadata = self.doc.metadata

        for page in self.doc:
            text.append(page.get_text())
            image_list = page.get_images(full=True)
            for img in image_list:
                xref = img[0]
                base_image = self.doc.extract_image(xref)
                image_bytes = base_image["image"]
                image = Image.open(io.BytesIO(image_bytes))
                images.append(image)

        self.text = "\n".join(text)
        self.images = images
        self.metadata = metadata
        return self.text, self.images, self.metadata

    def detect_jurisdiction(self, metadata, text):
        if metadata.get('producer') and 'Adobe' in metadata.get('producer'):
            if 'US' in text:
                return 'US'
            elif 'EU' in text:
                return 'EU'
            elif 'UK' in text:
                return 'UK'
            elif 'BR' in text:
                return 'BR'
            elif 'CN' in text:
                return 'CN'
            elif 'DE' in text:
                return 'DE'
            elif 'FR' in text:
                return 'FR'
            elif 'JP' in text:
                return 'JP'
            elif 'MS' in text:
                return 'MS'
            elif 'SG' in text:
                return 'SG'
        return super().detect_jurisdiction(metadata, text)

    def apply_rules(self, config, doc_type):
        return super().apply_rules(config, doc_type)