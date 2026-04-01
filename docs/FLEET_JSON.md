# Fleet JSON Export Contract

This document describes the `fleet-json` compute run export:

```bash
./oci-collector -profile DEFAULT compute -r -f fleet-json -o fleet.json
```

## Purpose

`fleet-json` emits a sample-compatible fleet payload for downstream ingestion while preserving existing compute `json`/`csv` formats.

It also persists run history at `.DATA/compute_fleet_state.json` to derive:
- `Previous_Status`
- `Status_Change_Signal`
- `Status_Changed_UTC`
- reboot history/evidence fields

## Top-Level Shape

```json
{
  "generatedAt": "RFC3339 timestamp",
  "generatedAtEpochMs": 0,
  "schemaVersion": 2,
  "source": {
    "type": "oci-sdk",
    "auth": "config | instance-principal",
    "profile": "effective profile",
    "customerStrategy": "tenancy",
    "regions": ["region list"]
  },
  "customers": [
    {
      "name": "tenancy name",
      "lastImport": 0,
      "instanceCount": 0,
      "scheduledCount": 0,
      "completedCount": 0,
      "changedCount": 0
    }
  ],
  "instances": []
}
```

## Instance Field Notes

- Core identity/inventory fields (`ID`, `Display_Name`, `Shape`, `State`, `Compartment_*`, `Tenant_ID`, `Region`, `freeformTags`, `definedTags`) are sourced from compute inventory data.
- Maintenance fields are enriched from OCI Compute maintenance APIs:
  - `Maintenance_*`
  - `Maintenance_Event_OCID`
  - `Maintenance_Event_Status`
- `Maintenance_Reboot_Due` is driven by instance reboot-due availability.
- `Live_Migration_Preference` is inferred from maintenance action and availability config.

## History and Status Derivation

- `uniqueKey = customerId + "_" + ID`.
- Each run compares current status against prior snapshot state.
- If status changed, `Status_Change_Signal` is set to `<Previous_Status> -> <reboot_status>` and `Status_Changed_UTC` updates.
- If no change, `Status_Change_Signal = "No Change"`.

## Time Formatting

- UTC fields use `YYYY-MM-DD HH:MM:SS`.
- IST/EDT style fields use `YYYY-MM-DD HH:MM`.
- Export timestamps:
  - `generatedAt`: RFC3339 UTC
  - `generatedAtEpochMs`: epoch milliseconds

## Failure Behavior

In `fleet-json` mode, maintenance enrichment is strict:
- if maintenance API collection fails, command returns an error and does not silently degrade.
