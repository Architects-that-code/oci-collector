# Development Guide

## Prerequisites

- Go toolchain compatible with `go.mod` (`go 1.24` currently declared).
- OCI credentials via config file profile or Instance Principal.

## Local Build

```bash
go build ./...
```

Build a local binary:

```bash
go build -o oci-collector .
```

## Test

```bash
go test ./... -v
```

Current state: only `limits` package has committed tests; most packages report `[no test files]`.

## Run Locally

```bash
./oci-collector
./oci-collector config -v
./oci-collector compute -r
```

## Packaging

Use Makefile targets for cross-platform binaries:

```bash
make all
```

Artifacts are written to `dist/`.

## Recommended Verification Sequence

1. `go fmt ./...`
2. `go build ./...`
3. `go test ./... -v`
4. Smoke run one or two critical commands against a safe tenancy/profile.

## Configuration Files

- Runtime config: `toolkit-config.yaml`
- Template: `sample-toolkit-config.yaml`

Keep sensitive config values out of commits.

## Documentation Maintenance

When command flags or runtime behavior change:

1. Update `docs/COMMANDS.md`.
2. Update `docs/ARCHITECTURE.md` for flow/component changes.
3. Update `docs/PROJECT_ANALYSIS.md` if the risk profile changes.
4. Update root `ReadMe.md` links or quick-start if onboarding changes.
