# manifest:
#   task_slug: revert-422-fallback
#   model: google-vertex-anthropic/claude-sonnet-4-6@default
#   estimated_tokens: 25000
#   output_dir: ./reports

# Plan: Revert temporary 422 rich-search fallback
# Iteration: 004
# Author: Vader (orchestrator)
# Date: 2026-03-09
# Builds on: hotfix commit cd8fc21

## Context

During local debugging, a temporary fallback was added to `internal/brave/client.go` in
`SearchWithRich()`:
- If `GET /res/v1/web/search?q=...&enable_rich_callback=1` returned HTTP 422,
  the code fell back to `Search(query)` and returned standard web results.

This patch was useful during diagnosis, but the actual root cause was **not code**:
`brave-search-app/.env` contained a bad Brave API key, which caused HTTP 422 on all
endpoints (standard, rich, and news). After correcting the local `.env` key and restarting
server, all endpoints returned 200 again.

Therefore, the fallback patch is unnecessary and should be reverted to restore the intended
product behavior.

## Objective

Revert the temporary 422 fallback logic from `SearchWithRich()` and return the codebase to the
pre-hotfix behavior:
- Non-200 response from rich-enabled `/web/search` should return an error again
- No other behavior changes
- Do NOT touch `.env`
- Do NOT alter any rich UI/template work from plans 001–003

## Rollback Strategy

This plan itself is the rollback. If the revert introduces any unexpected regression:
- reset only the working tree changes from this run
- compare against commit `cd8fc21`
- stop and report; do not attempt a second speculative fix in the same run

## IMPORTANT: Read each file before editing it

## Implementation Steps

### 1. Read `internal/brave/client.go`
Understand the current `SearchWithRich()` logic and identify the temporary 422 branch.

### 2. Revert only the temporary fallback block
Inside `SearchWithRich()`, change this:

```go
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnprocessableEntity {
			// Some Brave API tiers reject enable_rich_callback=1 with 422.
			// Gracefully fall back to standard web search instead of surfacing an error.
			results, err := c.Search(query)
			if err != nil {
				return nil, nil, err
			}
			return &Response{Web: WebData{Results: results}}, nil, nil
		}
		return nil, nil, fmt.Errorf("api returned non-200 status code: %d", resp.StatusCode)
	}
```

Back to:

```go
	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("api returned non-200 status code: %d", resp.StatusCode)
	}
```

No other changes.

### 3. Run verification
Run:
```bash
go test ./...
go vet ./...
gofmt -l .
```

### 4. Report completion
Summarize exactly what changed and confirm that the revert was isolated to `client.go`.

## File Change Summary

| File | Action |
|---|---|
| `internal/brave/client.go` | Remove temporary 422 fallback branch from `SearchWithRich()` |

## Constraints

- Read file before editing
- No changes outside `internal/brave/client.go`
- Do not touch `.env`
- Do not update README
- Do not introduce new fallback behavior
- Standard library only

## Success Criteria

- `internal/brave/client.go` matches intended pre-hotfix logic
- Tests pass
- Vet clean
- Formatting clean
- No unrelated file modifications
