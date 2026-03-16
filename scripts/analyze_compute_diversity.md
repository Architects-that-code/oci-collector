# `analyze_compute_diversity.py`

Last updated: 2026-03-16
Script: `scripts/analyze_compute_diversity.py`

## Purpose

Analyze OCI compute placement diversity and report only risk conditions.

Risk categories:

1. Similar-name placement concentration within a region.
2. Region-wide concentration where all instances share one AD/FD placement key.

## Input Contract

Expected input file: JSON array of objects.

Fields used:
- `region`
- `displayName`
- `availabilityDomain`
- `faultDomain`

Fallback placeholders when fields are missing/empty:
- `region`: `<missing-region>`
- `displayName`: `<missing-displayName>`
- `availabilityDomain`: `<missing-ad>`
- `faultDomain`: `<missing-fd>`

## Name Normalization

Grouping key uses this suffix removal regex:

```regex
-[a-z0-9]+-\d+$
```

Example:
- `myapp-worker-abc123-2` -> `myapp-worker`

## Risk Definitions

### 1) Similar-name risk

Within one `(region, normalized_name)` group:
- instance count `> 1`
- unique AD/FD keys `< instance count`

AD/FD key format:

```text
{availabilityDomain}_{faultDomain}
```

### 2) Region concentration risk

Within a region (all names combined):
- instance count `> 1`
- exactly one unique AD/FD key

## Usage

Text output:

```bash
python3 scripts/analyze_compute_diversity.py <input.json>
```

JSON output:

```bash
python3 scripts/analyze_compute_diversity.py <input.json> --json
```

Fail CI/job when risk exists:

```bash
python3 scripts/analyze_compute_diversity.py <input.json> --fail-on-risk
```

## Exit Codes

- `0`: success without `--fail-on-risk` violation
- `1`: file/read/parse/input validation error
- `2`: `--fail-on-risk` set and any risk exists

## Current Caveats

1. Missing AD/FD fields are replaced with placeholders, which can inflate risk counts.
2. Name normalization only handles the defined suffix pattern.
3. No committed automated tests found for this script.

## Suggested Next Improvements

1. Add automated tests for no-risk and each risk mode.
2. Add strict mode for required AD/FD fields.
3. Support CSV output for risk rows.
4. Add threshold flags (for example minimum group size) for noisy environments.
