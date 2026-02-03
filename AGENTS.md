# AI Agent Guide for Brave Search App (Go)

This document outlines the development, testing, and style guidelines for the Brave Search App. AI agents operating in this codebase should strictly adhere to these rules to ensure consistency and maintainability.

## 1. Project Context
- **Type:** Go Web Application (Native `net/http`)
- **Version Management:** `mise` (preferred)
- **Go Version:** Latest (managed via `mise.toml`)
- **Core Dependencies:** Standard Library only

## 2. Environment Setup
The project uses `mise` for tool versioning.

```bash
# Install tools defined in mise.toml
mise install

# Configuration is handled via environment variables
export BRAVE_API_KEY=your_api_key_here
```

## 3. Build, Lint, and Test Commands

### Running the Application
```bash
# Run the server directly
go run cmd/server/main.go
```

### Testing
The project uses the standard Go `testing` package.

- **Run all tests:**
  ```bash
  go test ./...
  ```

- **Run tests in a specific package:**
  ```bash
  go test ./internal/brave
  ```

- **Run a specific test case (IMPORTANT):**
  Use the `-run` flag followed by a regex of the test name.
  ```bash
  go test -v ./internal/brave -run TestClient_Search/success
  ```

- **Run tests with verbose output:**
  ```bash
  go test -v ./...
  ```

### Linting and Formatting
Adhere to the Google Go Style Guide.

- **Check formatting:**
  ```bash
  gofmt -l .
  ```

- **Run static analysis:**
  ```bash
  go vet ./...
  ```

## 4. Code Style Guidelines

### Go General
- **Formatting:** Strictly use `gofmt`.
- **Indentation:** Use tabs for indentation.
- **Line Length:** Aim for 80 characters, max 100.

### Imports
Group imports into three categories, separated by a blank line:
1. Standard library
2. Third-party libraries (if any)
3. Local project imports

**Example:**
```go
import (
	"fmt"
	"net/http"

	"github.com/johanesalxd/brave-search-app/internal/brave"
)
```

### Naming Conventions
- **Exported Symbols:** `PascalCase` (e.g., `SearchService`)
- **Private Symbols:** `camelCase` (e.g., `internalHelper`)
- **Acronyms:** Use all caps (e.g., `HTTPClient`, `URL`, `ID`)
- **Short Names:** Use concise names for short-lived variables (e.g., `r` for `*http.Request`).

### Error Handling
- **Explicit Checks:** Always check errors immediately. Never use `_` to ignore errors.
- **Context:** Wrap errors with context using `fmt.Errorf("...: %w", err)`.
- **Punctuation:** Error strings should be lower-case and have no trailing punctuation.

### Documentation
- **Comments:** Every exported symbol must have a doc comment.
- **Style:** Comments must be complete sentences starting with the symbol name.
- **Examples:** Include `ExampleXxx` functions in `_test.go` files for "testable documentation".

**Example:**
```go
// Search performs a web search using the Brave Search API.
func (c *Client) Search(query string) ([]Result, error) {
	...
}
```

### Testing Patterns
- **Table-Driven Tests:** Use struct slices for test cases and `t.Run()` for subtests.
- **Isolation:** Use `httptest` to mock external API calls.

### Frontend (Templates)
- HTML files reside in `templates/`.
- Use Go `html/template` syntax consistently (`{{ .Variable }}`, `{{ range .List }}`).
- Ensure static assets (CSS) are linked via `/static/`.

## 5. Agent Behavior Protocols

1.  **Safety First:** Never commit secrets. Ensure `.env` or sensitive keys are never tracked.
2.  **Standard Library First:** Favor the Go standard library over third-party dependencies.
3.  **Verification:** Always run `go test ./...` and `go vet ./...` before considering a task complete.
4.  **Documentation:** When adding new public functions, ensure they are documented and have an associated `Example` test if appropriate.
5.  **Self-Correction:** If a test fails, analyze the failure, fix the specific error, and verify with a targeted test run.

## 6. Git Workflow
- **Commit Messages:** Use imperative mood (e.g., "Add handler for search").
- **Atomic Commits:** Keep logic changes separate from documentation or formatting updates.
