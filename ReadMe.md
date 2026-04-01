# OCI Collector

OCI Collector is a Go-based CLI for tenancy-wide Oracle Cloud Infrastructure (OCI) visibility.
It is designed for repeated operational queries (limits, compute, IAM, network, support, billing, and related inventory) across subscribed regions and compartments.

This project is a convenience layer on top of OCI APIs and SDKs. It is not intended to replace the OCI CLI or OCI SDK.

## What You Can Do

- Inventory running compute instances and optionally export JSON/CSV or sample-compatible fleet JSON (`-f fleet-json`).
- Collect compute utilization metrics and run metric namespace discovery.
- Inspect service limits by region.
- Enumerate IAM users, groups, and policies.
- Inspect object storage, network assets, support tickets, cloud advisor recommendations, and more.
- Run organization and child-tenancy related queries.

## Quick Start

1. Build:

```bash
go build -o oci-collector .
```

2. Create `toolkit-config.yaml` from `sample-toolkit-config.yaml`.

3. Set one auth mode:
   - Config file mode: `configPath`, `profileName`, `useinstanceprincipal: false`
   - Instance Principal mode: `useinstanceprincipal: true`

4. Run a smoke check:

```bash
./oci-collector config -v
```

Optional: override the configured OCI profile for a single run:

```bash
./oci-collector -profile MY_PROFILE compute -r
```

## Minimal Config

```yaml
configPath: ~/.oci/config
profileName: DEFAULT
useinstanceprincipal: false
SUPPORT_CSI_NUMBER: "123456789"
```

## Command Categories

- Core inventory: `config`, `limits`, `compute`, `object`, `network`, `search`
- IAM: `peeps`, `groups`, `policies`
- Operations and advisory: `support`, `cloudadvisor`, `billing`, `schedule`
- Capacity and platform: `capacity`, `capability`, `autonomous`, `children`

For full command and flag reference, see [docs/COMMANDS.md](docs/COMMANDS.md).

## Documentation Index

- Architecture: [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)
- Commands and flags: [docs/COMMANDS.md](docs/COMMANDS.md)
- Fleet JSON contract: [docs/FLEET_JSON.md](docs/FLEET_JSON.md)
- Development workflow: [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md)
- Project analysis and roadmap: [docs/PROJECT_ANALYSIS.md](docs/PROJECT_ANALYSIS.md)
- Compute diversity script docs: [scripts/analyze_compute_diversity.md](scripts/analyze_compute_diversity.md)

## Build and Test

```bash
go test ./...
```

Cross-platform binaries are produced via:

```bash
make all
```

## Notes

- `go.mod` currently declares `go 1.24`.
- Most packages currently have no automated tests; `limits` includes test coverage.

## License and Warranty

See [LICENSE](LICENSE).
Oracle and affiliates disclaim warranties for material in this repository.
