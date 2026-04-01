# Command Reference

This document reflects command behavior in `cmd/*.go`.

## Global Invocation

```bash
./oci-collector <command> [flags]
```

If a command receives no required action flag, many commands print help or a guidance message.

Global flags:
- `-profile` or `--profile` override `profileName` from `toolkit-config.yaml` for the current run.

## Configuration and Discovery

### `config`
Prints subscribed regions, AD count, and compartment count.

Flags:
- `-v, --verbose` include detailed region, AD, and compartment names.

## Compute and Capacity

### `compute`
Inventory running instances, collect metrics, enable metrics agent integration, or discover working monitoring dimensions.

Flags:
- `-r, --run` run compute inventory collection.
- `-m, --metrics` collect CPU utilization summaries.
- `-v, --verbose` include per-instance detail in text mode.
- `--progress` show progress during inventory collection (default `true`).
- `--enable-metrics` enable Oracle Cloud Agent metric collection on instances.
- `--metrics-discover` probe monitoring namespace/dimension combinations.
- `--discover-window` lookback duration for discovery (default `1h`).
- `--discover-instance` optional instance OCID filter for discovery.
- `-f, --format` output format (metrics default `json`; run supports `json`, `csv`, or `fleet-json` when specified).
- `-o, --out` optional output file path.

`fleet-json` writes sample-compatible fleet export JSON and maintains local run history at `.DATA/compute_fleet_state.json` for status-change and reboot-history fields. (this is the same format used by the https://github.com/Architects-that-code/oci-tenancy-explorer project so you could generate the fleet export file with `oci-collector compute -r --fleet-json -o fleet_data.json` and then import it into that tool for analysis or visualization).

### `capacity`
Capacity checks by OCPU, memory, and shape/chipset family.

Flags:
- `-r, --run`
- `-o, --ocpus`
- `-m, --memory`
- `-t, --type` (shape or chipset family)

### `capability`
Reports platform support characteristics for a shape family.

Flags:
- `-r, --run`
- `-t, --type`

### `autonomous`
Lists autonomous databases across subscribed regions and compartments.

Flags:
- `-r, --run`

## IAM and Policy

### `peeps`
Lists users.

Flags:
- `-r, --run`

### `groups`
Lists groups.

Flags:
- `-r, --run`

### `policies`
Lists policies by compartment.

Flags:
- `-r, --run`
- `-v, --verbose`

## Limits, Network, and Object Storage

### `limits`
Collects limits across subscribed regions.

Flags:
- `-r, --run`
- `-w, --write` write outputs to file(s).

### `network`
Collects network inventory.

Flags:
- `-r, --run`
- `-c, --cidr` include CIDR details.
- `-i, --ip` include IP inventory.
- `--rpc` include remote peering connection data.

### `object`
Collects object storage info.

Flags:
- `-r, --run`

## Search, Support, Billing, Advisor, Scheduling

### `search`
Runs resource search with a required search string.

Flags:
- `-s, --searchstring`

### `support`
Lists support tickets (uses configured CSI).

Flags:
- `-l, --list` required action flag to list tickets.

### `billing`
Downloads and/or processes billing files.
Requires at least one action flag.

Flags:
- `-p, --path` output/report directory (default `reports`).
- `-d, --download` download billing report files.
- `-x, --process` process downloaded billing files.
- `-e, --redownload-errors` retry files in `error_files.txt`.

### `cloudadvisor`
Fetches home-region recommendations and optionally resource actions.

Flags:
- `-f, --format` `json` or `text`.
- `-o, --out` optional output file.
- `--actions` include actions payload.
- `--org` include organization child tenancies.
- `--child-tenancies` comma-separated child tenancy OCIDs.

### `children`
Child tenancy collection and optional output write.

Flags:
- `-r, --run`
- `-w, --write`

### `schedule`
Runs scheduler information path.

Flags:
- `-r, --run`

## Root Command

### `oci-collector`
Base command; prints help by default.

Flags:
- `-profile` or `--profile` override `profileName` from `toolkit-config.yaml`.
