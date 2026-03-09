# manifest:
#   task_slug: rich-search-integration
#   model: google-vertex-anthropic/claude-sonnet-4-6@default
#   estimated_tokens: 80000
#   output_dir: ./reports

# Plan: Rich Search Integration
# Iteration: 001
# Author: Vader (orchestrator)
# Date: 2026-03-09

## Context

The current app calls a single endpoint:
- `GET https://api.search.brave.com/res/v1/web/search?q=<query>`
- Returns: `web.results[]` with `title`, `url`, `description`
- Struct: `brave.Response{Web: WebData{Results: []Result}}`

The Rich Search API is a two-step process:
1. **Step 1:** Call web/search with `enable_rich_callback=1` added to the query params
2. **Step 2:** If response contains a `rich.hint.callback_key`, call `GET /res/v1/web/rich?callback_key=<key>`
3. Rich results have a `type` (always "rich") and `subtype` (vertical: weather, stock, calculator, currency, crypto, sports, etc.)

Rich Search requires the **Search plan** API key (same key, different plan tier). The app's existing `BRAVE_API_KEY` is used — no new auth needed.

## Objective

Add Rich Search support with adaptive search box UI:
- When a query triggers a rich result, display the rich widget above standard web results
- Search box UI adapts to show which mode is active (standard vs rich)
- No breaking changes to existing web search flow

---

## Implementation Plan

### 1. New types in `internal/brave/types.go`

Add the following structs alongside existing `Result`, `WebData`, `Response`:

```go
// RichHint is embedded in web search responses when rich data is available.
type RichHint struct {
    Vertical    string `json:"vertical"`
    CallbackKey string `json:"callback_key"`
}

// RichField appears in the web search Response when enable_rich_callback=1
// and the query has a matching rich vertical.
type RichField struct {
    Type string   `json:"type"`
    Hint RichHint `json:"hint"`
}

// RichResult is the response from the /web/rich endpoint.
type RichResult struct {
    Type    string          `json:"type"`
    Subtype string          `json:"subtype"`
    Data    json.RawMessage `json:"data"` // vertical-specific payload, passed as-is to template
}

// Response update — add Rich field:
// Replace existing Response struct with:
type Response struct {
    Web  WebData    `json:"web"`
    Rich *RichField `json:"rich,omitempty"`
}
```

### 2. Update `internal/brave/client.go`

**Add constants:**
```go
const (
    DefaultAPIBaseURL    = "https://api.search.brave.com/res/v1/web/search"
    RichAPIBaseURL       = "https://api.search.brave.com/res/v1/web/rich"
    DefaultTimeout       = 30 * time.Second
)
```

**Update `Search()` method signature and behavior:**
- Add `richEnabled bool` parameter OR detect automatically via new `SearchWithRich()` method
- Preferred: add a new method `SearchWithRich(query string) (*Response, error)` that:
  1. Calls web/search with `enable_rich_callback=1`
  2. If `response.Rich != nil`, calls `/web/rich?callback_key=<key>` and attaches result
  3. Returns the full `Response` (caller decides how to render)
- Keep existing `Search(query string) ([]Result, error)` unchanged for backward compat

**New method:**
```go
// SearchWithRich performs a web search with rich data enrichment enabled.
// If the query triggers a rich vertical, it fetches the rich result and
// attaches it to the returned Response.
func (c *Client) SearchWithRich(query string) (*Response, *RichResult, error)
```

**New private helper:**
```go
func (c *Client) fetchRich(callbackKey string) (*RichResult, error)
```

### 3. Update `internal/handlers/server.go`

**Add `SearchMode` concept:**
```go
type SearchMode string

const (
    ModeStandard SearchMode = "standard"
    ModeRich     SearchMode = "rich"
)
```

**Update `HandleSearch()`:**
- Read `mode` query param from URL (`?q=...&mode=rich` or `?q=...&mode=standard`)
- Default: `ModeRich` (try rich first, fall back to standard if no rich result returned)
- If `mode=standard`: call existing `client.Search()`
- If `mode=rich` (default): call `client.SearchWithRich()`
- Pass `SearchMode`, `RichResult`, and `Results` to template data struct

**Updated template data:**
```go
data := struct {
    Query      string
    Results    []brave.Result
    RichResult *brave.RichResult
    Mode       SearchMode
    HasRich    bool
}{...}
```

### 4. Update `templates/index.html`

Add mode toggle to search box:
- Two toggle buttons: "Standard" / "Rich" (default: Rich)
- Toggle sets a hidden `mode` form field
- Visual indicator (e.g., pill/badge) showing active mode
- Submit sends `?q=<query>&mode=<standard|rich>`

### 5. Update `templates/results.html`

Add rich widget section above standard results:
- If `{{ .HasRich }}`: render a `<div class="rich-widget">` with vertical label (`{{ .RichResult.Subtype }}`) and raw data preview
- Keep existing `{{ range .Results }}` block unchanged below
- Display current mode badge in results header: "Rich Search" or "Standard Search"

### 6. Add tests

**`internal/brave/client_test.go`:**
- `TestClient_SearchWithRich/rich_result_returned` — mock web/search returning `rich` field + mock rich endpoint returning RichResult
- `TestClient_SearchWithRich/no_rich_result` — mock web/search with no `rich` field, verify only web results returned
- `TestClient_SearchWithRich/rich_fetch_error` — mock rich endpoint returning 500, verify graceful fallback

**`internal/handlers/server_test.go`:**
- `TestServer_HandleSearch/mode_rich_with_rich_result`
- `TestServer_HandleSearch/mode_standard_bypasses_rich`
- `TestServer_HandleSearch/mode_rich_no_rich_result_falls_back`

---

## File Change Summary

| File | Action |
|---|---|
| `internal/brave/types.go` | Add `RichHint`, `RichField`, `RichResult`; update `Response` |
| `internal/brave/client.go` | Add `RichAPIBaseURL` const; add `SearchWithRich()` + `fetchRich()` |
| `internal/brave/client_test.go` | Add 3 new test cases for `SearchWithRich` |
| `internal/handlers/server.go` | Add `SearchMode`; update `HandleSearch()` data struct |
| `internal/handlers/server_test.go` | Add 3 new handler test cases |
| `templates/index.html` | Add mode toggle UI to search form |
| `templates/results.html` | Add rich widget section; add mode badge |

No new dependencies. Standard library only. Existing `Search()` method untouched.

---

## Constraints (from AGENTS.md)

- Standard library only — no external packages
- Run `go test ./...` and `go vet ./...` before marking complete
- All new exported symbols must have doc comments
- Error strings: lowercase, no trailing punctuation
- Wrap errors: `fmt.Errorf("context: %w", err)`
- Table-driven tests with `t.Run()` subtests

---

## Verification Steps

After implementation, run:
```bash
go vet ./...
go test ./...
gofmt -l .
```

Expected: all tests pass, no vet errors, no formatting issues.
