# manifest:
#   task_slug: rich-widget-ui
#   model: google-vertex-anthropic/claude-sonnet-4-6@default
#   estimated_tokens: 80000
#   output_dir: ./reports

# Plan: Rich Widget UI — Structured Vertical Cards
# Iteration: 003
# Author: Vader (orchestrator)
# Date: 2026-03-09
# Builds on: plan-002-news-search-integration.md

## Context

Current state of rich widget in `templates/results.html`:
```html
{{ if .HasRich }}
    <div class="rich-widget">
        <span class="rich-vertical-label">{{ .RichResult.Subtype }}</span>
        <pre class="rich-data">{{ printf "%s" .RichResult.Data }}</pre>
    </div>
{{ end }}
```

Problem: `RichResult.Data` is `json.RawMessage` — rendered as raw JSON in a `<pre>` block.
Unreadable for users. Needs structured per-vertical cards.

`RichResult` in `internal/brave/models.go`:
```go
type RichResult struct {
    Type    string          `json:"type"`
    Subtype string          `json:"subtype"`
    Data    json.RawMessage `json:"data"`
}
```

Approach: Parse `Data` into typed structs per vertical in the handler (Go side),
pass a `RichDisplay` interface to the template. Template switches on subtype to
render appropriate card. Unknown subtypes get a pretty-printed JSON fallback.

## Supported Verticals (priority order)

Handle these 6 first (highest query frequency):
1. **weather** — temp, condition, location, forecast
2. **stock** — symbol, price, change, change_percent
3. **cryptocurrency** — symbol, name, price_usd, change_24h
4. **currency** — from, to, rate, result
5. **calculator** — expression, result
6. **definitions** — word, definitions[]

All others (sports, unit_conversion, package_tracker, unix_timestamp) →
**fallback: pretty-printed JSON in styled card** (not raw `<pre>`).

---

## Implementation Plan

### IMPORTANT: Read each file before editing it.

### 1. Add per-vertical structs to `internal/brave/models.go`

Read `internal/brave/models.go` first, then append:

```go
// RichWeather holds weather data from the rich search vertical.
type RichWeather struct {
    Location    string  `json:"location"`
    TempC       float64 `json:"temp_c"`
    TempF       float64 `json:"temp_f"`
    Condition   string  `json:"condition"`
    Humidity    int     `json:"humidity"`
    WindKph     float64 `json:"wind_kph"`
    ForecastDay string  `json:"forecast_day"`
}

// RichStock holds stock market data from the rich search vertical.
type RichStock struct {
    Symbol        string  `json:"symbol"`
    Name          string  `json:"name"`
    Price         float64 `json:"price"`
    Change        float64 `json:"change"`
    ChangePercent float64 `json:"change_percent"`
    Currency      string  `json:"currency"`
}

// RichCrypto holds cryptocurrency data from the rich search vertical.
type RichCrypto struct {
    Symbol    string  `json:"symbol"`
    Name      string  `json:"name"`
    PriceUSD  float64 `json:"price_usd"`
    Change24h float64 `json:"change_24h"`
    MarketCap float64 `json:"market_cap"`
}

// RichCurrency holds currency conversion data from the rich search vertical.
type RichCurrency struct {
    FromCurrency string  `json:"from_currency"`
    ToCurrency   string  `json:"to_currency"`
    Rate         float64 `json:"rate"`
    Amount       float64 `json:"amount"`
    Result       float64 `json:"result"`
}

// RichCalculator holds calculator result data from the rich search vertical.
type RichCalculator struct {
    Expression string `json:"expression"`
    Result     string `json:"result"`
}

// RichDefinition holds a single word definition entry.
type RichDefinition struct {
    PartOfSpeech string `json:"part_of_speech"`
    Definition   string `json:"definition"`
    Example      string `json:"example"`
}

// RichDefinitions holds word definition data from the rich search vertical.
type RichDefinitions struct {
    Word        string           `json:"word"`
    Definitions []RichDefinition `json:"definitions"`
}

// RichDisplay is passed to the results template and holds the parsed rich data.
// Only one of the typed fields will be non-nil, matching the Subtype.
// FallbackJSON holds pretty-printed JSON for unknown subtypes.
type RichDisplay struct {
    Subtype     string
    Weather     *RichWeather
    Stock       *RichStock
    Crypto      *RichCrypto
    Currency    *RichCurrency
    Calculator  *RichCalculator
    Definitions *RichDefinitions
    FallbackJSON string // pretty-printed for unknown subtypes
}
```

Note: The Brave API's exact JSON field names for rich verticals are not fully
documented publicly. Use the struct field names as best-effort guesses based on
the documented vertical types. If unmarshaling fails for a vertical, fall back
to pretty-printed JSON gracefully — do NOT return an error.

### 2. Add `ParseRichDisplay` helper to `internal/brave/models.go`

```go
// ParseRichDisplay converts a RichResult into a RichDisplay suitable for
// template rendering. It parses the Data field into the appropriate typed
// struct based on Subtype. On any parse failure, it falls back to
// pretty-printed JSON. ParseRichDisplay never returns an error.
func ParseRichDisplay(r *RichResult) *RichDisplay {
    if r == nil {
        return nil
    }
    d := &RichDisplay{Subtype: r.Subtype}
    switch r.Subtype {
    case "weather":
        var v RichWeather
        if json.Unmarshal(r.Data, &v) == nil {
            d.Weather = &v
            return d
        }
    case "stock":
        var v RichStock
        if json.Unmarshal(r.Data, &v) == nil {
            d.Stock = &v
            return d
        }
    case "cryptocurrency":
        var v RichCrypto
        if json.Unmarshal(r.Data, &v) == nil {
            d.Crypto = &v
            return d
        }
    case "currency":
        var v RichCurrency
        if json.Unmarshal(r.Data, &v) == nil {
            d.Currency = &v
            return d
        }
    case "calculator":
        var v RichCalculator
        if json.Unmarshal(r.Data, &v) == nil {
            d.Calculator = &v
            return d
        }
    case "definitions":
        var v RichDefinitions
        if json.Unmarshal(r.Data, &v) == nil {
            d.Definitions = &v
            return d
        }
    }
    // Fallback: pretty-print raw JSON
    var pretty []byte
    if b, err := json.MarshalIndent(json.RawMessage(r.Data), "", "  "); err == nil {
        pretty = b
    } else {
        pretty = r.Data
    }
    d.FallbackJSON = string(pretty)
    return d
}
```

### 3. Update `internal/handlers/handlers.go`

Read `internal/handlers/handlers.go` first, then:

In `HandleSearch()`, update the data struct and rich result handling:

Replace:
```go
data := struct {
    Query      string
    Results    []brave.Result
    RichResult *brave.RichResult
    Mode       SearchMode
    HasRich    bool
}{...}
```

With:
```go
data := struct {
    Query       string
    Results     []brave.Result
    RichDisplay *brave.RichDisplay
    Mode        SearchMode
    HasRich     bool
}{...}
```

Where `RichDisplay` is set by calling `brave.ParseRichDisplay(richResult)` after
`SearchWithRich()` returns. Set `HasRich = data.RichDisplay != nil`.

### 4. Update `templates/results.html`

Read `templates/results.html` first, then replace the rich widget block:

```html
{{ if .HasRich }}
<div class="rich-widget rich-{{ .RichDisplay.Subtype }}">

    {{ if .RichDisplay.Weather }}
    <div class="rich-card rich-weather">
        <div class="rich-card-header">🌤 Weather — {{ .RichDisplay.Weather.Location }}</div>
        <div class="rich-card-body">
            <span class="rich-temp">{{ .RichDisplay.Weather.TempC }}°C / {{ .RichDisplay.Weather.TempF }}°F</span>
            <span class="rich-condition">{{ .RichDisplay.Weather.Condition }}</span>
            <span class="rich-meta">Humidity: {{ .RichDisplay.Weather.Humidity }}% · Wind: {{ .RichDisplay.Weather.WindKph }} km/h</span>
        </div>
    </div>

    {{ else if .RichDisplay.Stock }}
    <div class="rich-card rich-stock">
        <div class="rich-card-header">📈 {{ .RichDisplay.Stock.Symbol }} — {{ .RichDisplay.Stock.Name }}</div>
        <div class="rich-card-body">
            <span class="rich-price">{{ .RichDisplay.Stock.Currency }} {{ printf "%.2f" .RichDisplay.Stock.Price }}</span>
            <span class="rich-change {{ if lt .RichDisplay.Stock.Change 0.0 }}negative{{ else }}positive{{ end }}">
                {{ printf "%+.2f" .RichDisplay.Stock.Change }} ({{ printf "%+.2f" .RichDisplay.Stock.ChangePercent }}%)
            </span>
        </div>
    </div>

    {{ else if .RichDisplay.Crypto }}
    <div class="rich-card rich-crypto">
        <div class="rich-card-header">₿ {{ .RichDisplay.Crypto.Symbol }} — {{ .RichDisplay.Crypto.Name }}</div>
        <div class="rich-card-body">
            <span class="rich-price">USD {{ printf "%.2f" .RichDisplay.Crypto.PriceUSD }}</span>
            <span class="rich-change {{ if lt .RichDisplay.Crypto.Change24h 0.0 }}negative{{ else }}positive{{ end }}">
                24h: {{ printf "%+.2f" .RichDisplay.Crypto.Change24h }}%
            </span>
        </div>
    </div>

    {{ else if .RichDisplay.Currency }}
    <div class="rich-card rich-currency">
        <div class="rich-card-header">💱 Currency Conversion</div>
        <div class="rich-card-body">
            <span class="rich-conversion">
                {{ printf "%.2f" .RichDisplay.Currency.Amount }} {{ .RichDisplay.Currency.FromCurrency }}
                = {{ printf "%.4f" .RichDisplay.Currency.Result }} {{ .RichDisplay.Currency.ToCurrency }}
            </span>
            <span class="rich-meta">Rate: {{ printf "%.6f" .RichDisplay.Currency.Rate }}</span>
        </div>
    </div>

    {{ else if .RichDisplay.Calculator }}
    <div class="rich-card rich-calculator">
        <div class="rich-card-header">🧮 Calculator</div>
        <div class="rich-card-body">
            <span class="rich-expression">{{ .RichDisplay.Calculator.Expression }}</span>
            <span class="rich-result">= {{ .RichDisplay.Calculator.Result }}</span>
        </div>
    </div>

    {{ else if .RichDisplay.Definitions }}
    <div class="rich-card rich-definitions">
        <div class="rich-card-header">📖 {{ .RichDisplay.Definitions.Word }}</div>
        <div class="rich-card-body">
            {{ range .RichDisplay.Definitions.Definitions }}
            <div class="rich-definition">
                <span class="rich-pos">{{ .PartOfSpeech }}</span>
                <p>{{ .Definition }}</p>
                {{ if .Example }}<p class="rich-example"><em>{{ .Example }}</em></p>{{ end }}
            </div>
            {{ end }}
        </div>
    </div>

    {{ else }}
    <div class="rich-card rich-fallback">
        <div class="rich-card-header">{{ .RichDisplay.Subtype }}</div>
        <pre class="rich-json">{{ .RichDisplay.FallbackJSON }}</pre>
    </div>
    {{ end }}

</div>
{{ end }}
```

### 5. Add rich card CSS to `static/style.css`

Read `static/style.css` first, then append styles for:
- `.rich-widget` — container with border-radius, subtle background, margin-bottom
- `.rich-card` — card base: padding, border-left accent color per vertical
- `.rich-card-header` — bold label, small caps
- `.rich-card-body` — flex column, gap
- `.rich-temp`, `.rich-price`, `.rich-result` — large font for primary value
- `.rich-change.positive` — green color
- `.rich-change.negative` — red color
- `.rich-meta`, `.rich-pos` — muted small text
- `.rich-expression` — monospace
- `.rich-json` — monospace, small font, max-height 200px, overflow-y scroll
- Per-vertical accent: `.rich-weather` border-left blue, `.rich-stock` border-left green,
  `.rich-crypto` border-left orange, `.rich-currency` border-left purple,
  `.rich-calculator` border-left teal, `.rich-definitions` border-left indigo

### 6. Update tests

**`internal/brave/models_test.go`** (new file — read directory first to confirm it doesn't exist):
- `TestParseRichDisplay/weather` — unmarshal weather JSON, verify Weather field populated
- `TestParseRichDisplay/stock` — verify Stock field populated
- `TestParseRichDisplay/unknown_subtype` — verify FallbackJSON non-empty, typed fields nil
- `TestParseRichDisplay/nil_input` — verify nil returned

**`internal/handlers/handlers_test.go`** — Read first, then:
- Update `TestServer_HandleSearch/mode_rich_with_rich_result` to use `RichDisplay` field
  instead of `RichResult` in template assertions (update `resultsTemplate` const and assertions)

---

## File Change Summary

| File | Action |
|---|---|
| `internal/brave/models.go` | Add 7 typed structs + `RichDisplay` + `ParseRichDisplay()` |
| `internal/brave/models_test.go` | New file — 4 test cases for `ParseRichDisplay` |
| `internal/handlers/handlers.go` | Update `HandleSearch()` data struct: `RichResult` → `RichDisplay` via `ParseRichDisplay()` |
| `internal/handlers/handlers_test.go` | Update `resultsTemplate` + assertions for `RichDisplay` |
| `templates/results.html` | Replace raw `<pre>` with per-vertical card blocks |
| `static/style.css` | Append rich card CSS |

No new dependencies. Standard library only.

---

## Constraints (from AGENTS.md)

- **Read each file before editing it** — mandatory for all edits
- Standard library only — no external packages
- Run `go test ./...` and `go vet ./...` before marking complete
- All new exported symbols must have doc comments
- Error strings: lowercase, no trailing punctuation
- Wrap errors: `fmt.Errorf("context: %w", err)`
- Table-driven tests with `t.Run()` subtests
- `ParseRichDisplay` must NEVER return an error — always falls back gracefully

---

## Verification Steps

After implementation, run:
```bash
go vet ./...
go test ./...
gofmt -l .
```

Expected: all tests pass, no vet errors, no formatting issues.
