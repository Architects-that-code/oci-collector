# OCI Collector Architecture

## Overview

OCI Collector is a single binary CLI application written in Go.
Execution starts in `main.go`, prints a banner, then delegates to Cobra commands in `cmd/`.

## High-Level Design

1. CLI entrypoint:
- `main.go` calls `cmd.Execute()`.
- `cmd/root.go` registers subcommands and default help behavior.

2. Configuration and auth:
- `config/config.go` reads `toolkit-config.yaml`.
- `config.Prep` builds an OCI `ConfigurationProvider` using profile config or Instance Principal.

3. Shared tenancy setup:
- `config.CommonSetup` retrieves subscribed regions, compartments, availability domains, and home region.
- Most commands call this during startup.

4. Service modules:
- Each command delegates to a package (`compute/`, `limits/`, `iam/`, `networks/`, `objectstorage/`, etc.) that performs OCI SDK calls and output formatting.

5. Output:
- Predominantly stdout text output.
- Several commands support structured export (JSON/CSV and file output flags).

## Component Map

- CLI layer: `cmd/*.go`
- Config and environment: `config/config.go`
- Shared utilities and retry helpers: `util/*.go`
- Service modules:
  - `compute/`
  - `limits/`
  - `iam/`
  - `billing/`
  - `cloudadvisor/`
  - `networks/`
  - `objectstorage/`
  - `support/`
  - `search/`
  - `schedule/`
  - `capCheck/`, `capability/`, `autonomous/`, `childtenancies/`

## Data Flow

Typical command flow:

1. User runs `oci-collector <command> [flags]`.
2. Cobra parses command and flags.
3. Command loads `toolkit-config.yaml`.
4. Command builds OCI provider/client.
5. Optional shared setup resolves tenancy context (regions, compartments, ADs, home region).
6. Service package executes OCI API calls.
7. Results are printed and/or written to output files.

## Concurrency and Reliability Patterns

- Multi-region and multi-compartment scans are parallelized in service packages.
- Backoff/retry logic exists for throttling-sensitive paths (for example compute listing/metrics behavior and `util/backoff.go`).
- Commands that mutate client region typically create region-specific clients, or in config setup run region-switch operations sequentially where noted.

## Boundaries and Current Constraints

- Runtime model: CLI only; no web service layer.
- Persistence: no internal database found.
- State: ephemeral in-process collections and optional output files.
- Global command flags are minimal; command-level flags carry most behavior.

## Related Docs

- Root overview: [../ReadMe.md](../ReadMe.md)
- Command reference: [COMMANDS.md](COMMANDS.md)
- Engineering analysis: [PROJECT_ANALYSIS.md](PROJECT_ANALYSIS.md)
