import unittest
from kite.nlp import classifier, compliance, translation

class TestDocumentClassifier(unittest.TestCase):
    def test_classify_contract(self):
        text = "This is a contract document."
        doc_type = classifier.DocumentClassifier.classify(text)
        self.assertEqual(doc_type, 'contract')

    def test_classify_court_decision(self):
        text = "This is a court decision document."
        doc_type = classifier.DocumentClassifier.classify(text)
        self.assertEqual(doc_type, 'court_decision')

    def test_classify_statute(self):
        text = "This is a statute document."
        doc_type = classifier.DocumentClassifier.classify(text)
        self.assertEqual(doc_type, 'statute')

    def test_classify_unknown(self):
        text = "Random text without keywords."
        doc_type = classifier.DocumentClassifier.classify(text)
        self.assertEqual(doc_type, 'unknown')

class TestComplianceChecker(unittest.TestCase):
    def test_check_missing_fields(self):
        text = "This document lacks signatures."
        config = {
            'document_types': {
                'contract': {
                    'required_fields': ['signatures'],
                    'risk_flags': ['missing_signatures']
                }
            }
        }
        risks = compliance.ComplianceChecker.check(text, config, 'contract')
        self.assertIn('missing_signatures', risks)

class TestTranslator(unittest.TestCase):
    def test_translate(self):
        text = "Hello"
        translated = translation.Translator.translate(text, target_lang='fr')
        self.assertIsInstance(translated, str)

if __name__ == '__main__':
    unittest.main()