import unittest
from kite.extractors.base import BaseExtractor

class DummyExtractor(BaseExtractor):
    def extract_all(self):
        return ('sample text', ['img1'], {'producer': 'Dummy'})

class TestExtractors(unittest.TestCase):
    def test_base_extractor(self):
        extractor = DummyExtractor('dummy.pdf')
        text, images, metadata = extractor.extract_all()
        self.assertEqual(text, 'sample text')
        self.assertEqual(images, ['img1'])
        self.assertEqual(metadata, {'producer': 'Dummy'})

    def test_detect_jurisdiction(self):
        extractor = DummyExtractor('dummy.pdf')
        jurisdiction = extractor.detect_jurisdiction({}, 'This is a US contract')
        self.assertEqual(jurisdiction, 'US')
        jurisdiction = extractor.detect_jurisdiction({}, 'European Union regulation')
        self.assertEqual(jurisdiction, 'EU')
        jurisdiction = extractor.detect_jurisdiction({}, 'UK contract')
        self.assertEqual(jurisdiction, 'UK')
        jurisdiction = extractor.detect_jurisdiction({}, 'Brazilian contract')
        self.assertEqual(jurisdiction, 'BR')
        jurisdiction = extractor.detect_jurisdiction({}, 'Chinese contract')
        self.assertEqual(jurisdiction, 'CN')
        jurisdiction = extractor.detect_jurisdiction({}, 'German contract')
        self.assertEqual(jurisdiction, 'DE')
        jurisdiction = extractor.detect_jurisdiction({}, 'French contract')
        self.assertEqual(jurisdiction, 'FR')
        jurisdiction = extractor.detect_jurisdiction({}, 'Japanese contract')
        self.assertEqual(jurisdiction, 'JP')
        jurisdiction = extractor.detect_jurisdiction({}, 'Malaysian contract')
        self.assertEqual(jurisdiction, 'MS')
        jurisdiction = extractor.detect_jurisdiction({}, 'Singaporean contract')
        self.assertEqual(jurisdiction, 'SG')
        jurisdiction = extractor.detect_jurisdiction({}, 'Other text')
        self.assertEqual(jurisdiction, None)
        

if __name__ == '__main__':
    unittest.main()