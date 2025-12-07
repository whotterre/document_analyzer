# AI Document Summarization + Metadata Extraction Workflow

## Problem
Build a service that accepts a PDF or DOCX file, extracts text, sends it to an LLM on OpenRouter, and returns:
- A concise summary
- Detected document type (invoice, CV, report, letter, etc.)
- Extracted metadata (date, sender, total amount, etc.)

## Requirements

### 1. Upload Document
**Endpoint:** `POST /documents/upload`

- **Input:** Multipart form data with a file (PDF/DOCX, max 5MB).
- **Process:**
    - Validate file type and size.
    - Store the raw file in S3 or Minio.
    - Extract text content using a library (e.g., `ledongthuc/pdf` or `unidoc`).
    - Save extracted text and initial file metadata in the database.
- **Output:** JSON response with the created Document ID.

### 2. Analyze Document
**Endpoint:** `POST /documents/{id}/analyze`

- **Process:**
    - Retrieve the extracted text for the given Document ID.
    - Send the text to an LLM via OpenRouter (e.g., `gpt-4o-mini` or other available models).
    - Prompt the LLM to generate:
        - A concise summary.
        - Document type classification.
        - Key metadata extraction (JSON format preferred).
    - Save the LLM output (summary, type, attributes) to the database.
- **Output:** JSON response indicating success or the analysis result.

### 3. Get Document Details
**Endpoint:** `GET /documents/{id}`

- **Process:**
    - Fetch document record from the database.
- **Output:** Combined JSON response containing:
    - File information (original name, size, upload date).
    - Extracted text (optional or truncated).
    - Generated summary.
    - Detected document type.
    - Extracted metadata.

## Tech Stack
- **Language:** Go
- **Web Framework:** Gin (suggested)
- **Storage:** S3 / Minio (for files), PostgreSQL / SQLite (for metadata)
- **AI/LLM:** OpenRouter API
