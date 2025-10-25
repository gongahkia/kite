import spacy

class DocumentClassifier:
    _nlp = None

    @classmethod
    def _load_model(cls, lang='en'):
        if cls._nlp is None:
            try:
                if lang == 'en':
                    cls._nlp = spacy.load('en_core_web_sm')
                else:
                    cls._nlp = spacy.blank(lang)
            except Exception:
                cls._nlp = spacy.blank('en')
        return cls._nlp

    @classmethod
    def classify(cls, text, lang='en'):
        nlp = cls._load_model(lang)
        doc = nlp(text)
        text_lower = text.lower()
        if 'contract' in text_lower:
            return 'contract'
        elif 'court decision' in text_lower or 'judgment' in text_lower:
            return 'court_decision'
        elif 'statute' in text_lower or 'law' in text_lower:
            return 'statute'
        else:
            return 'unknown'