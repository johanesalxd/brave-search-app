# Brave Search App

A lightweight Go-based web application that provides a clean interface for searching the web using the Brave Search API.

## Features
- Minimalist search interface
- Fast results powered by Brave Search API
- Built with Go standard library (no external web frameworks)
- Easy configuration via environment variables

## Setup & Installation

### Prerequisites
- Go 1.24 or higher
- [mise](https://mise.jdx.dev/) (recommended for version management)

### Installation

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd brave-search-app
   ```

2. **Configure Environment:**
   Ensure your environment has the Brave Search API key. You can use a `.env` file (if using `mise` or a similar tool that loads it) or export it:
   ```bash
   export BRAVE_API_KEY=your_api_key_here
   ```

## Usage

### Running the Application
Start the server:
```bash
go run cmd/server/main.go
```
The app will be available at `http://127.0.0.1:5001`.

## Development & Quality Control

### Formatting & Linting
We follow the Google Go Style Guide.
```bash
# Check formatting
gofmt -l .

# Run static analysis
go vet ./...
```

## Testing
The project uses the standard Go `testing` package.

### Run all tests
```bash
go test ./...
```

### Run tests with verbose output
```bash
go test -v ./...
```

## Project Structure
Adhering to the standard Go layout:
- `cmd/server/`: Main application entry point.
- `internal/brave/`: Core logic for Brave Search API interaction.
- `internal/handlers/`: HTTP request handlers.
- `templates/`: HTML templates (Go `html/template`).
- `static/`: Static assets (CSS).
