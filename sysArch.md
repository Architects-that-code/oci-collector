# System Architecture

## Overview

The OCI Collector is a command-line interface (CLI) tool built in Go, designed to provide a "utility belt" for interacting with Oracle Cloud Infrastructure (OCI). It allows users to quickly gather information about their OCI tenancy, such as resource usage, service limits, user permissions, and more. The tool is structured in a modular way, with each package corresponding to a specific OCI service or a distinct piece of functionality.

The application is executed from the command line, where the user specifies a command (e.g., `limits`, `compute`, `policies`) and optional flags to control the behavior of the command. The tool then uses the OCI Go SDK to communicate with the OCI API and retrieve the requested information. The project follows a standard Go structure with all source code under the root directory, including cmd/ for CLI, util/ for helpers, and service-specific directories (e.g., limits/, compute/). The root directory contains the go.mod for dependency management, main.go as the entry point, and the Makefile for building cross-platform binaries.

## High-Level Architecture

The architecture is designed to be modular and self-contained, with the root directory containing the go.mod for dependency management, main.go as the entry point, and directories for services and utilities. The main entry point is main.go, which parses arguments using the flag package and routes to the appropriate service logic. The structure ensures easy building with `go build` and cross-compilation via the Makefile for multiple platforms (macOS, Linux, Windows).

1. **CLI Entrypoint (`main.go`)**: The main entry point of the application. It imports the cmd package and calls cmd.Execute() to start the Cobra CLI. This handles parsing and routing to subcommands.

2. **Configuration (`config` package)**: Handles the loading and management of OCI connection configuration. It reads the necessary credentials and settings from the user's OCI configuration file or uses instance principals. The package is imported as "config" in other modules.

3. **CLI Commands (`cmd` package)**: Uses Cobra for structured subcommands. The root command is in cmd/root.go, which defines the base command and adds all subcommands. Each subcommand is in cmd/<command>.go (e.g., cmd/limits.go), with a Run function that calls the corresponding service logic (e.g., limits.RunLimits). Flags are defined in each subcommand (e.g., --run for limits). This pattern makes it easy to add new commands by creating cmd/<new>.go and adding it to root.go.

4. **OCI Service Modules**: The core logic of the application is divided into several packages, each corresponding to a specific OCI service. These modules are responsible for making the actual API calls to OCI and processing the results. The main modules include:
    * `limits`: Fetches service limits using OCI Limits API.
    * `compute`: Queries for compute instances using OCI Compute API.
    * `iam`: Handles Identity and Access Management (users, groups, policies using OCI Identity API).
    * `billing`: Manages billing and cost analysis using OCI Billing API.
    * `networks`: Manages Virtual Cloud Networks (VCNs, subnets, gateways, routes, security lists using OCI Core and Network APIs).
    * `objectstorage`: Interacts with Object Storage (buckets, sizes using OCI Object Storage API).
    * `support`: Manages support tickets using OCI Support API.
    * `search`: Searches for resources using OCI Search API.
    * `schedule`: Deals with scheduling using OCI Resource Scheduler API.
    * `capability`: Checks hardware capabilities using OCI Core API.
    * `capCheck`: Checks capacity using OCI Limits and Compute APIs.
    * `childtenancies`: Manages child tenancies using OCI Tenant Manager API.

5. **Utility Package (`util`)**: Provides common helper functions used across different modules, such as printing banners, logging, and caching for regions/compartments to reduce API calls. Imported as "util" in other modules.

6. **OCI Go SDK**: The application relies on the official OCI Go SDK (v65) to interact with the OCI API. The SDK handles the complexities of authentication, request signing, and communication with the OCI endpoints.

The root directory contains the `go.mod` file, `Makefile` for building, and any non-source files (e.g., ReadMe.md, sample-toolkit-config.yaml). The structure is flat, with all Go code in the root for simplicity, ensuring compliance with Go practices. No nested src/ dir; all Go code is directly under the root.

## Workflow

A typical workflow of the application is as follows:

1. The user executes the application from the command line, providing a command and any necessary flags (e.g., `go run main.go limits --run`).
2. The `main` function in `main.go` parses the command-line arguments.
3. Based on the command, the `switch` statement in `main` directs the execution to the corresponding service module (e.g., the `limits` package).
4. The `config` package is used to load the OCI configuration and prepare the OCI client.
5. The service module (e.g., `limits.RunLimits`) uses the OCI Go SDK to make API calls to the relevant OCI service.
6. The data received from the OCI API is processed and displayed to the user in the console or written to files.

### Detailed Data Flow Example (Limits Command)
1. Parse "limits --run" -> Call limits.RunLimits(provider, regions, tenancyID, write).
2. For each region (goroutine): ListServices -> For each service: ListLimitValues -> For each limit: GetResourceAvailability (region/AD scoped).
3. Collect LimitsCollector structs -> Sort/output console -> If write: Marshal JSON/YAML to files.
Similar patterns in other modules, with concurrency for multi-region queries and caching for common resources.

## Optimizations and Refactoring Suggestions

To improve maintainability, scalability, performance, and developer experience, consider the following:

### 1. **CLI Structure (High Priority)**
- **Refactor to Cobra/Urflag**: Replace manual flag parsing in main.go with Cobra for subcommands, auto-help, and validation. E.g., `go run main.go limits run --write`. Reduces boilerplate and improves usability.
- **Global Flags**: Add profile/region filters as global flags to avoid per-command repetition.

### 2. **Configuration & Setup (Medium Priority)**
- **Validate Config**: Add schema validation (e.g., using gojsonschema) for toolkit-config.yaml; handle missing fields gracefully.
- **Caching Common Resources**: Memoize regions/compartments (e.g., using sync.RWMutex) to reduce API calls by ~70% in multi-region scenarios.

### 3. **Performance & Concurrency (High Priority)**
- **Rate Limiting/Retries**: Commands like network hit 429 errors in multi-region queries. Implement exponential backoff (e.g., github.com/cenkalti/backoff) for API calls; limit concurrent calls to 5 per service to respect OCI quotas.
- **Bounded Concurrency**: Use errgroup for worker pools to limit parallel API calls, preventing overload.
- **Pagination Optimization**: Use SDK's pagination; cache paginated results to avoid re-fetching.
- **Async Outputs**: For large datasets (e.g., billing CSVs), stream to files concurrently to reduce memory usage.

### 4. **Error Handling & Observability (Medium Priority)**
- **Structured Logging**: Replace fmt.Printf with zap/slog; log at levels (info/debug/error); include context (region/comp).
- **Graceful Errors**: Use structured errors (errors package); provide user-friendly messages for rate limits (e.g., "Retrying in 5s...") and auth issues (e.g., "Check OCI config").

### 5. **Code Quality & Testing (High Priority)**
- **Unit/Integration Tests**: Add go test coverage; mock OCI SDK (e.g., using testify); test edge cases (empty regions, errors).
- **Linting/Formatting**: Enforce gofmt, golangci-lint; remove dead code (e.g., unused funcs like GetCompartmentsHeirarchy).
- **Type Safety**: Wrap OCI types in custom structs for better error handling and serialization (e.g., JSON output).

### 6. **Output & Extensibility (Low Priority)**
- **Consistent Formats**: Support JSON/YAML/CSV via flag; use templates (text/template) for custom outputs.
- **File Management**: Centralize writes (e.g., util.WriteOutput with timestamps/rotations).
- **Modularity**: Extract shared pagination logic; use interfaces for services (e.g., Limiter interface) for mocking and pluggable extensions (e.g., new OCI services).

### 7. **Security & Compliance (Low Priority)**
- **Secrets Handling**: Use env vars for sensitive config; validate OCI config.
- **Audit Logging**: Log API calls with user/context for compliance.

### Implementation Priority
- **High**: Rate limiting/retries, bounded concurrency, caching (performance in multi-region).
- **Medium**: Structured logging, unit tests, linting (developer usability).
- **Low**: Consistent formats, modularity, security enhancements (expansion).

These changes would make the tool more robust, easier to extend (e.g., new OCI services), and suitable for production use, with all source code under the root directory.