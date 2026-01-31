# AI Agent Guide for Brave Search App

This document outlines the development, testing, and style guidelines for the Brave Search App. AI agents operating in this codebase should strictly adhere to these rules to ensure consistency and maintainability.

## 1. Project Context
- **Type:** Python Web Application (Flask)
- **Package Manager:** `uv` (preferred) or `pip`
- **Python Version:** >= 3.13
- **Core Dependencies:** `flask`, `requests`, `python-dotenv`

## 2. Environment Setup
Before making changes, ensure dependencies are installed.

```bash
# Using uv (Recommended)
uv sync

# Using pip
pip install -r requirements.txt
```

## 3. Build, Lint, and Test Commands

### Running the Application
```bash
# Development mode (runs on port 5001)
python app.py
# OR
flask run --debug --port 5001
```

### Testing
*Note: The project uses `pytest` conventions, though explicit test files may need to be created in a `tests/` directory.*

- **Run all tests:**
  ```bash
  python -m pytest
  ```

- **Run a single test file:**
  ```bash
  python -m pytest tests/test_app.py
  ```

- **Run a specific test case (IMPORTANT):**
  When debugging or iterating on a specific feature, always run only the relevant test to save time.
  ```bash
  python -m pytest tests/test_app.py::test_search_functionality
  ```

- **Run tests with output:**
  Use `-s` to see print statements and `-v` for verbose output.
  ```bash
  python -m pytest -vs tests/test_app.py
  ```

### Linting and Formatting
Code should be linted using `ruff` (or `flake8`) and formatted with `black` compatible styles.

- **Check for linting errors:**
  ```bash
  ruff check .
  ```

- **Fix linting errors automatically:**
  ```bash
  ruff check . --fix
  ```

- **Format code:**
  ```bash
  ruff format .
  ```

## 4. Code Style Guidelines

### Python General
- **Standard:** Follow **PEP 8** strictly.
- **Indentation:** Use 4 spaces. No tabs.
- **Line Length:** 88 characters (Black standard) or 100 characters.

### Imports
Group imports in the following order, separated by a blank line:
1.  **Standard Library** (e.g., `os`, `sys`, `typing`)
2.  **Third-Party Libraries** (e.g., `flask`, `requests`)
3.  **Local Application Imports** (e.g., `from brave_search import search_brave`)

**Example:**
```python
import os
from typing import Optional, Dict

import requests
from flask import Flask

from brave_search import search_brave
```

### Naming Conventions
- **Variables/Functions:** `snake_case` (e.g., `user_query`, `fetch_results`)
- **Classes:** `PascalCase` (e.g., `SearchService`)
- **Constants:** `UPPER_CASE` (e.g., `BRAVE_API_KEY`, `MAX_RETRIES`)
- **Private Members:** Prefix with underscore (e.g., `_internal_helper`)

### Type Hinting
- **Strong Recommendation:** Add type hints to all new function signatures.
- Use `typing` module for complex types (though standard collection types `list`, `dict` are preferred in Python 3.9+).

**Example:**
```python
def search_brave(query: str) -> dict:
    ...
```

### Error Handling
- Use specific exception blocks (e.g., `except ValueError`) rather than bare `except:`.
- Propagate exceptions when the current scope cannot handle them meaningfully.
- For HTTP requests, always check response status (e.g., `response.raise_for_status()`).

### Documentation
- Add docstrings to all public modules, classes, and functions.
- Use **Google Style** docstrings.

**Example:**
```python
def search_brave(query: str) -> dict:
    """
    Executes a search query against the Brave Search API.

    Args:
        query (str): The search string.

    Returns:
        dict: The JSON response from the API.

    Raises:
        ValueError: If the API key is missing.
        requests.exceptions.HTTPError: If the API request fails.
    """
    ...
```

### Frontend (Templates)
- HTML files reside in `templates/`.
- Use Jinja2 syntax consistently (`{{ variable }}`, `{% block content %}`).
- Ensure static assets (CSS, JS) are linked correctly via `url_for('static', filename='...')`.

## 5. Agent Behavior Protocols

1.  **Safety First:** Never commit secrets (like `BRAVE_API_KEY`). Ensure `.env` is in `.gitignore`.
2.  **Incremental Changes:** When refactoring, make small, testable changes.
3.  **Verification:** Always run the relevant test *before* and *after* making changes to ensure no regression.
4.  **Self-Correction:** If a build or test fails, analyze the output, fix the specific error, and retry. Do not blindly attempt multiple random fixes.
5.  **New Dependencies:** If adding a new library, update `pyproject.toml` (or `requirements.txt`) immediately.

## 6. Git Workflow
- **Commit Messages:** Use imperative mood, concise and descriptive.
  - Good: "Add error handling for empty search queries"
  - Bad: "Fixed bug"
- **Atomic Commits:** Separate logic changes from formatting changes.
