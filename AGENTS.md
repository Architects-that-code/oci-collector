## Brief Overview

Project-specific workflow for OCI Collector development: compile, test, smoke run, and keep docs aligned with command behavior.

## Development Workflow

- After meaningful code edits, run `go build ./...`.
- Run tests with `go test ./... -v`.
- Run formatting: `go fmt ./...`.
- Keep dependencies clean with `go mod tidy` when needed.
- Run `golangci-lint run` if available.
- Build cross-platform artifacts with `make all` when preparing release-quality output.

## Smoke Checks

- `go run .`
- `go run . config -v`
- `go run . compute -r`
- Optional metrics path: `go run . compute --metrics`

## Documentation Sync Rules

When command behavior changes:

1. Update `docs/COMMANDS.md`.
2. Update `docs/ARCHITECTURE.md` when data flow/components change.
3. Update `docs/PROJECT_ANALYSIS.md` if risk profile or roadmap priorities change.
4. Ensure root `ReadMe.md` links stay valid.
5. Always include the Disclaimer (Sample Code) at the bottom of the ReadMe.

## Coding Guidance

- Prefer shared retry helpers in `util/backoff.go` for throttling paths.
- Keep user-facing flag behavior explicit and consistent.
- Remove dead flags or wire them to real behavior.

## Communication Style

- Be direct and technical.
- Report what was changed, what was verified, and what remains unverified.
- Use bullet points for clarity and brevity.