# Repository Guidelines

## Project Structure & Module Organization
- `api/api.yaml` is the canonical contract; update it before wiring new routes.
- Generated request/response bindings live in `internal/api`; `main.go` holds the `go:generate` directive for refreshes.
- HTTP orchestration logic belongs under `internal/handler` (endpoint handlers) and `internal/server` (router setup, middleware).
- Binary entry points reside in `cmd/<service>/main.go`; use descriptive folder names per deployable.
- Co-locate fixtures and test doubles with the packages they validate; add a shared `internal/testutil` package if reuse emerges.

## Build, Test, and Development Commands
- `go build ./...` compiles every package; run it before pushing to catch type breaks.
- `go test ./... -count=1` executes the suite without cache so flaky cases surface early.
- `go run ./cmd/<service>` boots a local binary; supply environment variables via your shell when needed.
- `go generate ./...` refreshes OpenAPI bindings after editing `api/api.yaml`.
- `go vet ./...` surfaces static analysis issues; treat warnings as blockers before review.

## Coding Style & Naming Conventions
- Target Go 1.25.1 and format code with `gofmt`/`goimports`; CI assumes clean diffs.
- Prefer table-driven tests and constructor helpers when registering routers or middleware.
- Name packages with short lowercase nouns (`server`, `handler`); export symbols only when cross-package use is required.
- Accept `context.Context` as the first argument for request-scope functions; favor interfaces at boundaries.
- Keep files focused (~300 lines max) and split routers by domain (`internal/handler/conversation`, `internal/handler/session`).

## Testing Guidelines
- Place `_test.go` files beside the code under test; mirror directory layout to aid discovery.
- Cover new endpoints with `httptest` request/response assertions and table-driven cases.
- Mock upstream services using lightweight fakes; store shared helpers in `internal/testutil` when added.
- Aim for substantial coverage on routing, serialization, and error paths; document intentional gaps in the PR body.

## Commit & Pull Request Guidelines
- Follow the existing `<type>: <summary>` pattern (`feat:`, `fix:`, `refactor:`) with imperative summaries.
- Scope commits narrowly and include regenerated code whenever the spec changes.
- PR descriptions should note intent, major code paths touched, and verification steps (tests, manual checks).
- Attach screenshots or `curl` transcripts when altering HTTP behavior to ease review.
- Request a peer review before merging and rebase on `main` to keep history linear.
