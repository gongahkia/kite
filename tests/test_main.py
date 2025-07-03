import unittest
from unittest.mock import patch
from kite.main import extract_document

class TestMainExtractionPipeline(unittest.TestCase):
    @patch('kite.extractors.pdf_extractor.PDFExtractor.extract_all')
    @patch('kite.extractors.pdf_extractor.PDFExtractor.detect_jurisdiction')
    @patch('kite.extractors.pdf_extractor.PDFExtractor.apply_rules')
    @patch('kite.nlp.classifier.DocumentClassifier.classify')
    @patch('kite.nlp.compliance.ComplianceChecker.check')
    @patch('kite.nlp.translation.Translator.translate')
    def test_extract_document(self, mock_translate, mock_check, mock_classify, mock_apply_rules, mock_detect_jurisdiction, mock_extract_all):
        mock_extract_all.return_value = ('sample text', ['image1', 'image2'], {'producer': 'Adobe'})
        mock_detect_jurisdiction.return_value = 'US'
        mock_classify.return_value = 'contract'
        mock_apply_rules.return_value = {'party_names': 'Party A and Party B'}
        mock_check.return_value = ['missing_signatures']
        mock_translate.return_value = 'texte exemple'

        result = extract_document('dummy.pdf', lang='en', translate_to='fr')

        self.assertEqual(result['jurisdiction'], 'US')
        self.assertEqual(result['document_type'], 'contract')
        self.assertIn('party_names', result['extracted_data'])
        self.assertIn('missing_signatures', result['risks'])
        self.assertEqual(result['text'], 'texte exemple')
        self.assertEqual(result['metadata'], {'producer': 'Adobe'})

if __name__ == '__main__':
    unittest.main()