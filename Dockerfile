FROM python:3.10-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        tesseract-ocr \
        libtesseract-dev \
        poppler-utils \
        build-essential \
        libglib2.0-0 \
        libsm6 \
        libxext6 \
        libxrender-dev \
        libmagic1 \
        && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY requirements.txt .
RUN pip install --upgrade pip && pip install -r requirements.txt

COPY src/ ./src/
COPY examples/ ./examples/
COPY tests/ ./tests/
COPY asset/ ./asset/
COPY pyproject.toml .
COPY setup.py .
COPY . .

EXPOSE 5000

CMD ["python", "-m", "src.api"]