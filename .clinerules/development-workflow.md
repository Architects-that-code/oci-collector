## Brief overview
Project-specific guidelines for OCI Collector CLI development, covering compilation, runnability, commit workflow, and tool usage.

## Development workflow
- After every replace_in_file or write_to_file, run `go build ./...` to verify no compile errors.
- Fix compilation errors before further changes.
- Test key commands: `go run . compute -r`, `go run . compute --metrics`.
- Run tests: `go test ./... -v`.
- Format: `go fmt ./...` and `go mod tidy`.
- Lint: `golangci-lint run` if available.
- Run `make all` before commits for cross-platform builds.
- Ensure project runs correctly: `go run .` (or subcommands).
- Fix races/panics immediately (e.g., shared state).
- Update task_progress after verification.
- Use final_file_content for subsequent SEARCH blocks.
- Limit <5 SEARCH/REPLACE blocks per call.

## Coding best practices
- Match SEARCH exactly (whitespace, indentation).
- Use one tool per message; wait for success.
- Add OCI backoff/retry.
- Verify with runnability/compilation checks before commit.

## Communication style
- Confirm "Build succeeds with `make all`" in attempt_completion.
- Be direct/technical; use checklists.