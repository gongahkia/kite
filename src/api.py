from flask import Flask, request, jsonify
from main import extract_document

app = Flask(__name__)

@app.route('/extract', methods=['POST'])
def extract():
    file = request.files['file']
    lang = request.form.get('lang', 'auto')
    translate_to = request.form.get('translate_to')
    file_path = f"/tmp/{file.filename}"
    file.save(file_path)
    result = extract_document(file_path, lang=lang, translate_to=translate_to)
    return jsonify(result)

def run():
    app.run(host='0.0.0.0', port=5000)