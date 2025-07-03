from googletrans import Translator as GoogleTranslator

class Translator:
    _translator = GoogleTranslator()

    @classmethod
    def translate(cls, text, target_lang='en'):
        if not text:
            return text
        try:
            translated = cls._translator.translate(text, dest=target_lang)
            return translated.text
        except Exception:
            return text