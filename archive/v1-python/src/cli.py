import argparse
from main import extract_document

def main():
    parser = argparse.ArgumentParser(description="KITE Legal Document Extractor")
    parser.add_argument("file", help="Path to the legal document")
    parser.add_argument("--lang", default="auto", help="Document language")
    parser.add_argument("--translate-to", help="Translate extracted text")
    args = parser.parse_args()
    result = extract_document(args.file, lang=args.lang, translate_to=args.translate_to)
    print(result)

if __name__ == "__main__":
    main()