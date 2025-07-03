import unittest
import argparse
from unittest.mock import patch

class TestCLI(unittest.TestCase):
    @patch('builtins.print')
    @patch('kite.main.extract_document')
    def test_cli_main(self, mock_extract_document, mock_print):
        mock_extract_document.return_value = {
            'jurisdiction': 'US',
            'document_type': 'contract',
            'extracted_data': {},
            'images': [],
            'risks': [],
            'text': '',
            'metadata': {}
        }

        parser = argparse.ArgumentParser(description="KITE Legal Document Extractor")
        parser.add_argument("file", help="Path to the legal document")
        parser.add_argument("--lang", default="auto", help="Document language")
        parser.add_argument("--translate-to", help="Translate extracted text")

        args = parser.parse_args(['dummy.pdf', '--lang', 'en', '--translate-to', 'fr'])

        from kite.cli import main
        main()

        mock_extract_document.assert_called_once_with('dummy.pdf', lang='en', translate_to='fr')

if __name__ == '__main__':
    unittest.main()