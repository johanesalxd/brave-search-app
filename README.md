# Brave Search App

A lightweight Go-based web application that provides a clean search interface powered by the Brave Search API. Supports standard web search, rich vertical cards, and news search — all with zero external dependencies.

Built iteratively with OpenClaw orchestration and an OpenCode coding agent via the `.opencode/plans/` workflow.

## Features

- **Web Search** — Standard web results via Brave Search API
- **Web+ (Rich Search)** — Enriched results with structured vertical cards for weather, stocks, crypto, currency, calculator, and word definitions
- **News Search** — Dedicated news results with source, age, and description
- **Three-tab UI** — Web / Web+ / News toggle on the search and results pages
- **Per-vertical rich cards** — Typed display (not raw JSON) with colour-coded CSS cards per vertical; unknown verticals fall back to pretty-printed JSON
- Built with Go standard library only — no external web frameworks or dependencies

## Setup & Installation

### Prerequisites

- Go 1.24 or higher
- A [Brave Search API](https://api-dashboard.search.brave.com/) key (free tier available)
- [mise](https://mise.jdx.dev/) (optional, recommended for Go version management)

### Installation

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd brave-search-app
   ```

2. **Trust and install tools (if using mise):**
   ```bash
   mise trust
   mise install
   ```

3. **Configure environment:**
   Create a `.env` file in the root directory:
   ```env
   BRAVE_API_KEY=your_api_key_here
   ```
   If using `mise`, it will automatically load this file on `go run`.

## Usage

### Running the Application

```bash
go run cmd/server/main.go
```

The app will be available at `http://127.0.0.1:5001`.

### Search Modes

| Mode | URL | Description |
|---|---|---|
| Web | `/search?q=<query>&mode=standard` | Standard web results |
| Web+ | `/search?q=<query>&mode=rich` | Rich vertical cards + web results |
| News | `/news?q=<query>` | News results with source + age |

Web+ is the default mode when no `mode` parameter is provided.

### Rich Verticals

When Brave returns a rich result for a query, Web+ mode renders a structured card above the web results:

| Vertical | Card shows |
|---|---|
| `weather` | Location, temp (°C/°F), condition, humidity, wind |
| `stock` | Symbol, name, price, change, change % |
| `cryptocurrency` | Symbol, name, USD price, 24h change |
| `currency` | From/to, rate, converted amount |
| `calculator` | Expression, result |
| `definitions` | Word, part of speech, definition, example |
| *(other)* | Pretty-printed JSON fallback card |

## Development & Quality Control

### Formatting & Linting

```bash
# Check formatting
gofmt -l .

# Run static analysis
go vet ./...
```

### Testing

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Force fresh run (no cache)
go test -count=1 ./...
```

## Project Structure

Standard Go layout:

```
brave-search-app/
├── cmd/server/          # Main application entry point
├── internal/
│   ├── brave/           # Brave Search API client + types
│   │   ├── client.go    # Search(), SearchWithRich(), NewsSearch()
│   │   ├── models.go    # Response types + ParseRichDisplay()
│   │   ├── client_test.go
│   │   └── models_test.go
│   └── handlers/        # HTTP request handlers
│       ├── handlers.go  # HandleIndex, HandleSearch, HandleNews
│       └── handlers_test.go
├── templates/           # Go html/template files
│   ├── index.html       # Search home (Web/Web+/News tabs)
│   ├── results.html     # Web/Web+ results + rich card blocks
│   └── news.html        # News results
├── static/
│   └── style.css        # Styles including rich card CSS
├── .opencode/plans/     # Iteration plans (opencode-wrapper convention)
└── reports/             # run-manifest.json (execution provenance)
```

## API Reference

Brave Search API endpoints used:

| Endpoint | Used for |
|---|---|
| `GET /res/v1/web/search` | Standard + Rich search (with `enable_rich_callback=1`) |
| `GET /res/v1/web/rich` | Rich vertical data fetch (callback_key from web search) |
| `GET /res/v1/news/search` | News search |

Authentication: `X-Subscription-Token` header.
