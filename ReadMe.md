# OCI Collector - A Utility Belt for OCI Tenancy Management

This project is a collection of Go-based CLI tools to help manage and monitor your OCI tenancy. It provides quick insights into limits, compute instances, users, policies, regions, compartments, object storage, support tickets, and more. Designed as a "Batman's utility belt" for common OCI queries, it's not a replacement for the OCI CLI or SDK but a convenient set of commands for frequent tasks.

**Key Features**:
- Global view of OCI estate across subscribed regions and compartments.
- Supports authentication via OCI config file or Instance Principal.
- Output in text (default), JSON, or CSV formats.
- Concurrency for faster multi-region queries with backoff for API throttling.

## Authentication
- **Config File**: Specify `configPath` and `profileName` in `toolkit-config.yaml` (default: `~/.oci/config`).
- **Instance Principal**: Set `useinstanceprincipal: true` in `toolkit-config.yaml`. Ensure proper policies and dynamic groups for the instance.
- Example `toolkit-config.yaml`:
  ```
  configPath: ~/.oci/config
  profileName: DEFAULT
  useinstanceprincipal: false
  SUPPORT_CSI_NUMBER: "123456789"  # Optional for support tickets
  ```

## Installation
1. Clone the repo: `git clone https://github.com/Architects-that-code/oci-collector.git`
2. Build: `go build -o oci-collector .`
3. Run: `./oci-collector <command> [flags]`

Or use `go run . <command>` for development.

## Commands

### Core Blocking and Tackling
- **limits**: View service limits across regions.
  - Example: `oci-collector limits -f json -o limits.json`
- **compute**: Fetch active compute instances (with backoff for throttling).
  - Flags: `-r/--run` (list instances), `-f/--format` (text/json/csv), `-o/--out` (file).
  - Example: `oci-collector compute -r -f csv -o instances.csv`
- **users** (peeps): List users and groups.
  - Example: `oci-collector peeps`
- **policies**: List IAM policies.
  - Example: `oci-collector policies`
- **regions**: List subscribed regions and compartments.
  - Example: `oci-collector config` (shows setup including regions/compartments).
- **shapes** (capacity): Where can I launch specific shapes? OS support by shape.
  - Example: `oci-collector capacity`
- **object**: List Object Storage buckets.
  - Example: `oci-collector object`

### Support and Advisor
- **support**: List open support, CAM, and account tickets.
  - Example: `oci-collector support`
- **cloudadvisor**: Collect Optimizer recommendations (home region only).
  - Flags: `-f/--format` (json/text).
  - Example: `oci-collector cloudadvisor -f json`

### Other Tools
- **autonomous**: List Autonomous Databases (ATP/ADW/AJD).
  - Example: `oci-collector autonomous`
- **billing**: View billing data.
  - Example: `oci-collector billing`
- **childtenancies**: List child tenancies.
  - Example: `oci-collector children`
- **groups**: List IAM groups.
  - Example: `oci-collector groups`
- **networks**: List VNICs and networks.
  - Example: `oci-collector network`
- **schedule**: Resource scheduler info.
  - Example: `oci-collector schedule`
- **search**: Resource search across tenancy.
  - Example: `oci-collector search`

## Architecture (sysArch.md Summary)
The project uses:
- **Cobra** for CLI structure (root + subcommands).
- **OCI Go SDK v65** for API calls with ConfigurationProvider.
- **Concurrency**: Goroutines for multi-region queries (e.g., in compute.RunCompute).
- **Backoff**: Exponential retry with jitter for API throttling (TooManyRequests/429).
- **Output**: Human-readable text, JSON, CSV via flags.
- **Config**: YAML-based with support for Instance Principal.

For detailed architecture, see [sysArch.md](sysArch.md).

## Development
- **Build**: `go build ./...`
- **Test**: `go test ./... -v` (focus on limits; add more!).
- **Format/Lint**: `go fmt ./... && go mod tidy && golangci-lint run`
- **Run with Verbose**: Many commands support verbose logging.

See [.clinerules](.clinerules) for pre-commit checks.

## Future Ideas
- Create support/CAM tickets via CLI.
- Output to email or custom metrics.
- Dry-run mode.
- Filtering by region/compartment/shape.
- Integrate with Terraform or Kubernetes for automation.

## License
See LICENSE. Oracle disclaims warranties; use at your own risk.

For contributions or issues, open a PR or ticket.