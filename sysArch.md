# System Architecture

## Table of Contents
- [Overview](#overview)
- [High-Level Architecture](#high-level-architecture)
- [Workflow](#workflow)
- [Optimizations and Refactoring Suggestions](#optimizations-and-refactoring-suggestions)
- [Future Improvements](#future-improvements)

## Overview

This project is a command-line interface (CLI) tool built in Go, designed to provide a "utility belt" for interacting with Oracle Cloud Infrastructure (OCI). It allows users to quickly gather information about their OCI tenancy, such as resource usage, service limits, user permissions, and more. The tool is structured in a modular way, with each package corresponding to a specific OCI service or a distinct piece of functionality.

The application is executed from the command line, where the user specifies a command (e.g., `limits`, `compute`, `policies`) and optional flags to control the behavior of the command. The tool then uses the OCI Go SDK to communicate with the OCI API and retrieve the requested information.

### Code Analysis Summary
- **Language & Dependencies**: Primarily Go 1.x with OCI Go SDK (v65). Key deps: yaml.v2 for config, slices/strings for utils. No external testing frameworks visible.
- **Entry Point**: main.go uses flag package for CLI parsing and routing via switch on os.Args[1]. Supports 13+ commands (e.g., limits, compute, billing).
- **Configuration**: toolkit-config.yaml loads profile, path, instance principal, CSI. Supports OCI config files or instance principal auth.
- **Common Setup**: config/CommonSetup fetches subscribed regions, compartments (hierarchical), ADs/FDs concurrently where possible.
- **Modules**:
  - **Billing**: Downloads/processes Cost Analysis CSVs via API; concurrency for downloads; outputs summaries/usage to files.
  - **Compute**: Lists running instances, VNICs, IPs across regions/comps; region-parallel; outputs shapes, tags, IPs.
  - **IAM (peopleresource)**: Lists users, groups, policies (with statements); comp-level for policies; basic pagination.
  - **Networks**: Fetches VCNs, subnets, gateways, routes, security lists, IP inventory; region-parallel; optional CIDR/IP details.
  - **Limits**: Queries service limits/values/availabilities (region/AD-scoped); region-parallel; outputs JSON/YAML if flagged.
  - **Object Storage**: Lists buckets, approx object counts/sizes; region-parallel; namespace per tenancy.
  - **Child Tenancies**: Lists child tenancies/details via Organization API; home-region focused.
  - **Others**: Support (tickets via CSI), Capacity/Capability (AD/FD checks for shapes/OCPUs/memory), Search (resource search by string), Schedule (basic scheduler stub).
- **Data Flows**: Auth -> Common resources (regions/comps/ADs) -> Per-command: Iterate regions/comps, paginate API calls, collect structs/slices, output console/files. Concurrency limited to regions; error handling via helpers.FatalIfError (panics on err).
- **Output/Storage**: Mostly console printf; some JSON/YAML writes (limits, billing); no DB/integration.
- **Strengths**: Modular by service; leverages OCI SDK; basic parallelism reduces latency for multi-region.
- **Weaknesses**: Tight coupling to OCI types; verbose output; inconsistent error handling; no caching/validation; potential rate limit issues without retries/backoff.

## High-Level Architecture

The architecture can be broken down into the following key components:

1.  **CLI Entrypoint (`main.go`):** This is the main entry point of the application. It is responsible for:
    *   Parsing command-line arguments using the `flag` package.
    *   Routing the execution to the appropriate package based on the user's command.
    *   Initializing the OCI configuration and session.

2.  **Configuration (`config` package):** This package handles the loading and management of OCI connection configuration. It reads the necessary credentials and settings from the user's OCI configuration file.

3.  **OCI Service Modules:** The core logic of the application is divided into several packages, each corresponding to a specific OCI service. These modules are responsible for making the actual API calls to OCI and processing the results. The main modules include:
    *   `billing`: Manages billing and cost analysis.
    *   `compute`: Queries for compute instances.
    *   `iam`: Handles Identity and Access Management (users, groups, policies).
    *   `limits`: Fetches service limits.
    *   `networks`: Manages Virtual Cloud Networks (VCNs).
    *   `objectstorage`: Interacts with Object Storage.
    *   `support`: Manages support tickets.
    *   `search`: Searches for resources.
    *   `schedule`: Deals with scheduling.
    *   `capability`: Checks for hardware capabilities.
    *   `capCheck`: Checks for capacity.
    *   `childtenancies`: Manages child tenancies.

4.  **OCI Go SDK:** The application relies on the official OCI Go SDK to interact with the OCI API. The SDK handles the complexities of authentication, request signing, and communication with the OCI endpoints.

5.  **Utility Package (`util`):** This package provides common helper functions that are used across different modules, such as printing banners and spacing to format the output.

### System Interactions
- **Authentication Flow**: Config loads YAML -> Prep() creates provider (file/instance principal) -> IdentityClient for tenancy/regions.
- **Resource Discovery**: CommonSetup() concurrently fetches regions (ListRegionSubscriptions), compartments (ListCompartments recursive), ADs (ListAvailabilityDomains per region).
- **Query Pattern**: For each service: SetRegion per region -> Paginate List* APIs (e.g., ListInstances, ListBuckets) -> Process/Output. Some use channels for regional aggregation.
- **Concurrency Model**: Goroutines for regions (e.g., in limits, compute); sync.WaitGroup + channels for collection. No worker pools or context cancellation.
- **Error Handling**: Relies on helpers.FatalIfError (os.Exit on err); no retries, logging minimal.
- **Data Structures**: OCI types directly (e.g., []identity.Compartment); custom structs in billing/limits (e.g., LimitsCollector, ReportSummary).

## Workflow

A typical workflow of the application is as follows:

1.  The user executes the application from the command line, providing a command and any necessary flags (e.g., `go run main.go limits -run`).
2.  The `main` function in `main.go` parses the command-line arguments.
3.  Based on the command, the `switch` statement in `main` directs the execution to the corresponding service module (e.g., the `limits` package).
4.  The `config` package is used to load the OCI configuration and prepare the OCI client.
5.  The service module (e.g., `limits.RunLimits`) uses the OCI Go SDK to make API calls to the relevant OCI service.
6.  The data received from the OCI API is processed and displayed to the user in the console.

### Detailed Data Flow Example (Limits Command)
1. Parse "limits -run" -> Call limits.RunLimits(provider, regions, tenancyID, write).
2. For each region (goroutine): ListServices -> For each service: ListLimitValues -> For each limit: GetResourceAvailability (region/AD scoped).
3. Collect LimitsCollector structs -> Sort/output console -> If write: Marshal JSON/YAML to files.
Similar patterns in other modules, with variations (e.g., billing downloads CSVs concurrently).

## Optimizations and Refactoring Suggestions

To improve maintainability, scalability, performance, and developer experience, consider the following:

### 1. **CLI Structure**
- **Refactor to Cobra/Urflag**: Replace manual flag/switch with Cobra for subcommands, auto-help, and validation. E.g., `oci-collector limits run --write`. Reduces boilerplate in main.go.
- **Global Flags**: Add profile/region filters as global flags to avoid per-command repetition.

### 2. **Configuration & Setup**
- **Validate Config**: Add schema validation (e.g., using gojsonschema) for toolkit-config.yaml; handle missing fields gracefully.
- **Caching Common Resources**: Memoize regions/compartments/ADs (e.g., using sync.Once or in-memory cache) since they're fetched repeatedly.
- **Dependency Injection**: Pass clients/interfaces to modules instead of recreating per call; use a factory for service clients.

### 3. **Performance & Concurrency**
- **Full Parallelism**: Extend goroutines to compartments/sub-resources; use worker pools (e.g., errgroup) for bounded concurrency to avoid overwhelming OCI APIs.
- **Rate Limiting/Retries**: Integrate backoff (e.g., github.com/cenkalti/backoff) for API calls; respect OCI quotas.
- **Pagination Optimization**: Use higher limits where possible; parallelize pagination if API supports.
- **Async Outputs**: For large datasets (e.g., billing CSVs), stream to files concurrently.

### 4. **Error Handling & Observability**
- **Structured Logging**: Replace fmt.Printf with zap/slog; log at levels (info/debug/error); include context (region/comp).
- **Graceful Errors**: Avoid FatalIfError; return errors up the chain, provide user-friendly messages (e.g., "Failed to fetch limits for region X: reason").
- **Metrics/Tracing**: Add Prometheus metrics for API calls/latency; OpenTelemetry for distributed tracing.

### 5. **Code Quality & Testing**
- **Interfaces/Abstract**: Define service interfaces (e.g., Limiter { GetLimits() }) for mocking in tests.
- **Unit/Integration Tests**: Add go test coverage; mock OCI SDK (e.g., using testify); test edge cases (empty regions, errors).
- **Linting/Formatting**: Enforce gofmt, golangci-lint; remove dead code (e.g., unused funcs like GetCompartmentsHeirarchy).
- **Type Safety**: Reduce nil checks; use structs over raw OCI types for custom data (e.g., expand LimitsCollector).

### 6. **Output & Extensibility**
- **Consistent Formats**: Default to JSON/YAML/CSV via flag; use templates (text/template) for custom outputs.
- **File Management**: Centralize writes (e.g., util.WriteOutput(path, data, format)); add timestamps/rotations.
- **Integrations**: Add hooks for email (net/smtp), metrics (OCI Monitoring), or webhooks; support dry-run mode.
- **Modularity**: Extract shared pagination logic; make modules pluggable (e.g., plugin system for new services).

### 7. **Security & Compliance**
- **Secrets Handling**: Avoid hardcoding; use env vars for sensitive config.
- **Audit Logging**: Log all API calls with user/context for compliance.

### Implementation Priority
- High: Cobra refactor, error handling, tests (quick wins for DX).
- Medium: Concurrency improvements, logging (performance/observability).
- Low: Integrations, advanced caching (feature expansion).

These changes would make the tool more robust, easier to extend (e.g., new OCI services), and suitable for production use.

## Future Improvements

The `ReadMe.md` file suggests some potential future improvements, including:

*   Refactoring the CLI to use a more robust framework like [Cobra](https://github.com/spf13/cobra).
*   Adding the ability to create support tickets.
*   Outputting data to files in various formats.
*   Sending output as email.
*   Publishing data as custom metrics.