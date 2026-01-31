# Brave Search App

A lightweight Flask-based web application that provides a clean interface for searching the web using the Brave Search API.

## Features
- Minimalist search interface
- Fast results powered by Brave Search API
- Easy configuration via environment variables

## Setup & Installation

### Prerequisites
- Python 3.13 or higher
- [uv](https://github.com/astral-sh/uv) (recommended) or pip

### Installation

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd brave-search-app
   ```

2. **Install dependencies:**
   Using `uv`:
   ```bash
   uv sync
   ```
   Using `pip`:
   ```bash
   pip install -r requirements.txt
   ```

3. **Configure Environment:**
   Create a `.env` file in the root directory and add your Brave Search API key:
   ```env
   BRAVE_API_KEY=your_api_key_here
   ```

## Usage

### Running the Application
Start the development server:
```bash
# Option 1: Run directly with Python
python app.py

# Option 2: Run using Flask CLI
flask run --debug --port 5001
```
The app will be available at `http://127.0.0.1:5001`.

## Development & Quality Control

### Linting & Formatting
We use `ruff` for linting and formatting.
```bash
# Check for errors
ruff check .

# Format code
ruff format .
```

## Testing

The project uses `pytest` for testing.

### Run all tests
```bash
python -m pytest
```

### Run tests with coverage output
```bash
python -m pytest -v
```

## Project Structure
- `app.py`: Main Flask application routes.
- `brave_search.py`: Core logic for interacting with Brave Search API.
- `templates/`: HTML templates (Jinja2).
- `static/`: Static assets (CSS/JS).
- `tests/`: Unit and integration tests.
