# Project Analysis (Current State)

Analysis date: 2026-03-16
Scope: repository source and command paths in `cmd/`, service packages, config, scripts, and build/test files.

## Executive Summary

OCI Collector is a practical multi-command OCI operations CLI with broad tenancy coverage and strong utility value for read-heavy operational workflows.
The architecture is straightforward and maintainable at small-to-medium scale, but test coverage and interface consistency should be improved before treating it as production-grade automation tooling.

## What Is Working Well

1. Clear modular separation by service domain.
- Service code is organized into package directories (`compute`, `limits`, `iam`, `networks`, etc.).
- Cobra command handlers in `cmd/` provide predictable entry points.

2. Useful tenancy-wide workflow.
- Shared setup path (`config.CommonSetup`) centralizes regions/compartments/home-region discovery.
- Most commands follow a consistent prep pattern: load config -> prepare provider/client -> gather tenancy context -> execute service logic.

3. Real operational features.
- Compute inventory and metric workflows (`run`, `metrics`, `metrics-discover`, `enable-metrics`).
- Cloud Advisor and child-tenancy support.
- Billing download/process flow with re-download support.

4. Reliability primitives are present.
- Backoff/retry patterns for throttling-sensitive OCI calls.
- Utility helpers in `util/backoff.go` and service-specific retry logic.

## Architecture Observations

- Runtime model is CLI-only; no API server or UI.
- No internal persistence store found.
- Output model is stdout-first, with selective JSON/CSV export support.
- Command implementations depend directly on OCI SDK client types, which is practical but limits testability without mocks/interfaces.

## Risk and Gap Analysis

### High Priority

1. Limited automated test coverage.
- `go test ./...` shows test files only in `limits`.
- Regressions in command wiring and service logic are likely to escape until runtime.

2. Command behavior inconsistencies.
- Some commands define flags not currently enforced or consumed (`support --list`, `capacity --ad/--fd`).
- Behavioral contract can be ambiguous for operators and automation scripts.

3. Mixed retry strategy placement.
- Retry/backoff exists in both shared utility and local service implementations.
- Inconsistent usage can produce uneven behavior across commands.

### Medium Priority

1. Config and secret handling ergonomics.
- Runtime requires local YAML file; no schema validation observed.
- Inline comments/values in local `toolkit-config.yaml` indicate operational drift risk.

2. Output contract consistency.
- Some commands expose file/format flags; others are text-only.
- Machine-consumable automation is stronger where JSON output is standardized.

3. User-facing command ergonomics.
- Several commands rely on `-r/--run` switches while others execute immediately.
- Global conventions are not uniform.

### Low Priority

1. Documentation drift risk.
- Prior markdown had stale guidance and merge artifacts.
- Reorganized docs fix this now, but process checks are still needed.

2. Technical debt signals.
- Empty or placeholder methods exist (for example `GetCompartmentsHeirarchy`).
- Historical binaries and generated artifacts in repo root can increase noise.

## Recommended Roadmap

### Phase 1: Stability and Contracts

1. Standardize command behavior contract:
- decide which commands require explicit `--run`
- remove or implement currently-unused flags
- ensure help text reflects real behavior

2. Create baseline tests:
- unit tests for command flag handling and output mode switches
- table tests for utility retry helpers
- targeted integration-style tests around compute metrics formatting and search input validation

3. Add config validation:
- validate required fields for each auth mode before making API calls
- produce explicit actionable error messages

### Phase 2: Maintainability

1. Consolidate retry strategy to shared utility.
2. Introduce interfaces around OCI client operations where tests need mocking.
3. Normalize output format support across major commands.

### Phase 3: Operator Experience

1. Add deterministic JSON schema docs for machine-consumed commands.
2. Provide command cookbook examples for common runbooks.
3. Add CI checks for docs drift (command help snapshot or command metadata generation).

## Evidence Snapshot

- Commands discovered in `cmd/root.go`: 17 subcommands.
- Go tests run successfully on 2026-03-16 (`go test ./...`), with test files present only in `limits`.
- Build and release model present in `Makefile` with cross-platform targets into `dist/`.

## Related Docs

- Architecture: [ARCHITECTURE.md](ARCHITECTURE.md)
- Commands: [COMMANDS.md](COMMANDS.md)
- Development: [DEVELOPMENT.md](DEVELOPMENT.md)
