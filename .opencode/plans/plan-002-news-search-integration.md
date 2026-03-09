# manifest:
#   task_slug: news-search-integration
#   model: google-vertex-anthropic/claude-sonnet-4-6@default
#   estimated_tokens: 70000
#   output_dir: ./reports

# Plan: News Search Integration
# Iteration: 002
# Author: Vader (orchestrator)
# Date: 2026-03-09
# Builds on: plan-001-rich-search-integration.md

## Context

Current app state (post plan-001):
- Routes: `GET /` (index), `GET /search` (web + rich search)
- `internal/brave/client.go`: `Search()`, `SearchWithRich()`, `fetchRich()`
- `internal/brave/types.go`: `Result`, `WebData`, `Response`, `RichHint`, `RichField`, `RichResult`
- `internal/handlers/server.go`: `SearchMode` (standard/rich), `HandleIndex`, `HandleSearch`
- `cmd/server/main.go`: registers `/` and `/search` routes only

News Search API endpoint (GET):
- URL: `GET https://api.search.brave.com/res/v1/news/search?q=<query>`
- Auth: `X-Subscription-Token: <API_KEY>` (same key)
- Key params: `q` (required), `count` (max 50, default 20), `offset` (0-based, max 9),
  `freshness` (pd=24h / pw=7d / pm=31d / py=1yr / custom YYYY-MM-DDtoYYYY-MM-DD),
  `country` (2-char), `search_lang`, `ui_lang`, `safesearch` (off/moderate/strict, default strict),
  `extra_snippets` (bool), `spellcheck` (bool)

Note on GET vs POST: GET chosen — single-query search box use case, RESTful, cacheable,
consistent with existing web search pattern. POST is for bulk/batch payloads, not applicable here.

## Objective

Add a News Search tab/mode to the app:
- New `/news` route handles news queries
- Search box gains a "News" toggle alongside existing "Web" (standard/rich)
- News results show title, URL, description, source name, and article age
- No breaking changes to existing `/search` route or web/rich flow

---

## Implementation Plan

### 1. New types in `internal/brave/types.go`

Add alongside existing types:

```go
// NewsResult represents a single news article from the Brave News Search API.
type NewsResult struct {
    Title       string `json:"title"`
    URL         string `json:"url"`
    Description string `json:"description"`
    Age         string `json:"age"`         // e.g. "2 hours ago" or ISO timestamp
    Source      string `json:"meta_url"`    // source domain/name, nested in API response
}

// NewsResponse represents the top-level structure of the Brave News Search API response.
type NewsResponse struct {
    Results []NewsResult `json:"results"`
}
```

Note: The Brave News API returns results under a top-level `results` array (not nested under
a `news` key like web search nests under `web`). Verify exact field names by reading the
actual API response structure. If `meta_url` is a nested object, extract the `hostname` field
or define a sub-struct and implement custom JSON unmarshaling. Keep it simple — if nested,
use a flat struct with `json:"-"` and a custom `UnmarshalJSON`.

### 2. New method in `internal/brave/client.go`

Add constant:
```go
// NewsAPIBaseURL is the production URL for the Brave News Search API.
NewsAPIBaseURL = "https://api.search.brave.com/res/v1/news/search"
```

Add `newsURL` field to `Client` struct. Update `NewClient()` to set it.
Update `NewClientWithURLs()` to accept a third `newsURL string` parameter.

New method:
```go
// NewsSearch performs a news search using the Brave News Search API.
// It returns a list of news results or an error if the request fails.
func (c *Client) NewsSearch(query string) ([]NewsResult, error)
```

Implementation mirrors existing `Search()`:
1. Validate query non-empty
2. Parse `c.newsURL`, set `q` param
3. GET with `Accept: application/json` and `X-Subscription-Token` headers
4. Decode into `NewsResponse`, return `resp.Results`

### 3. Update `internal/handlers/server.go`

Add new handler:
```go
// HandleNews processes news search queries and renders news results.
func (s *Server) HandleNews(w http.ResponseWriter, r *http.Request)
```

Implementation:
1. Read `q` param — redirect to `/` if empty
2. Call `s.braveClient.NewsSearch(query)`
3. Render `news.html` with data struct:

```go
data := struct {
    Query   string
    Results []brave.NewsResult
}{
    Query:   query,
    Results: results,
}
```

### 4. Update `cmd/server/main.go`

Register new route:
```go
mux.HandleFunc("/news", server.HandleNews)
```

### 5. Update `templates/index.html`

Add search type selector above/beside the search box:
- Three options: "Web" (standard), "Web+" (rich, current default), "News"
- Implementation: radio buttons or tab links that set `action="/search"` or `action="/news"`
- "Web+" is the existing rich mode (default), "Web" is standard mode
- "News" routes to `/news`
- Keep existing hidden `mode` field logic for Web/Web+ toggle
- Visual: simple pill-style toggle, consistent with existing UI

### 6. New `templates/news.html`

New template for news results page:
- Show query at top with active "News" tab highlighted
- For each result: title (linked), source + age on same line, description below
- No rich widget (news has no rich enrichment)
- Back-to-search link at top

Structure:
```html
{{ range .Results }}
<article class="news-result">
  <h3><a href="{{ .URL }}">{{ .Title }}</a></h3>
  <p class="meta">{{ .Source }} &middot; {{ .Age }}</p>
  <p class="description">{{ .Description }}</p>
</article>
{{ end }}
```

---

## File Change Summary

| File | Action |
|---|---|
| `internal/brave/types.go` | Add `NewsResult`, `NewsResponse` |
| `internal/brave/client.go` | Add `NewsAPIBaseURL` const; add `newsURL` to Client; update `NewClient` + `NewClientWithURLs`; add `NewsSearch()` |
| `internal/brave/client_test.go` | Add `TestClient_NewsSearch` with 3 subtests (success, empty query, api error) |
| `internal/handlers/server.go` | Add `HandleNews()` |
| `internal/handlers/handlers_test.go` | Add `TestServer_HandleNews` with 2 subtests (success, empty query redirect) |
| `cmd/server/main.go` | Register `/news` route |
| `templates/index.html` | Add Web/Web+/News search type toggle |
| `templates/news.html` | New template for news results |

No new dependencies. Standard library only.

---

## Constraints (from AGENTS.md)

- Standard library only — no external packages
- Run `go test ./...` and `go vet ./...` before marking complete
- All new exported symbols must have doc comments
- Error strings: lowercase, no trailing punctuation
- Wrap errors: `fmt.Errorf("context: %w", err)`
- Table-driven tests with `t.Run()` subtests
- `NewClientWithURLs` signature change: add `newsURL string` as third param — update existing
  call sites in `client_test.go` and `handlers_test.go` accordingly

---

## Verification Steps

After implementation, run:
```bash
go vet ./...
go test ./...
gofmt -l .
```

Expected: all tests pass, no vet errors, no formatting issues.
