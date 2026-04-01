package compute

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/oracle/oci-go-sdk/v65/common"
	core65 "github.com/oracle/oci-go-sdk/v65/core"
)

func TestBuildFleetPayloadEmptyInventoryHasContractKeys(t *testing.T) {
	snapshotPath := filepath.Join(t.TempDir(), "compute_fleet_state.json")
	payload, err := BuildFleetPayload(nil, nil, "ocid1.tenancy.oc1..example", nil, RunMetadata{
		AuthType:         "config",
		Profile:          "DEFAULT",
		CustomerStrategy: "tenancy",
		TenancyName:      "Example Tenancy",
		SnapshotPath:     snapshotPath,
	})
	if err != nil {
		t.Fatalf("BuildFleetPayload error: %v", err)
	}

	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	for _, k := range []string{"generatedAt", "generatedAtEpochMs", "schemaVersion", "source", "customers", "instances"} {
		if _, ok := m[k]; !ok {
			t.Fatalf("missing key %q in payload", k)
		}
	}
}

func TestNormalizeRebootStatus(t *testing.T) {
	if got := normalizeRebootStatus(false, core65.InstanceMaintenanceEventSummary{}, false); got != "Not Scheduled" {
		t.Fatalf("expected Not Scheduled, got %q", got)
	}
	if got := normalizeRebootStatus(false, core65.InstanceMaintenanceEventSummary{}, true); got != "Scheduled" {
		t.Fatalf("expected Scheduled from reboot due, got %q", got)
	}

	ev := core65.InstanceMaintenanceEventSummary{LifecycleState: core65.InstanceMaintenanceEventLifecycleStateSucceeded}
	if got := normalizeRebootStatus(true, ev, false); got != "Completed" {
		t.Fatalf("expected Completed, got %q", got)
	}

	ev = core65.InstanceMaintenanceEventSummary{LifecycleState: core65.InstanceMaintenanceEventLifecycleStateProcessing}
	if got := normalizeRebootStatus(true, ev, false); got != "Scheduled" {
		t.Fatalf("expected Scheduled, got %q", got)
	}
}

func TestMaintenanceTimestampFields(t *testing.T) {
	ts := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
	ev := core65.InstanceMaintenanceEventSummary{
		TimeWindowStart: &common.SDKTime{Time: ts},
	}
	utc, ist, edt := maintenanceTimestampFields(true, ev, nil)
	if utc != "2026-01-15 12:00:00" {
		t.Fatalf("unexpected UTC %q", utc)
	}
	if ist != "2026-01-15 17:30" {
		t.Fatalf("unexpected IST %q", ist)
	}
	if edt != "2026-01-15 07:00" && edt != "2026-01-15 08:00" {
		t.Fatalf("unexpected EDT/EST %q", edt)
	}
}

func TestSnapshotRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	in := fleetSnapshot{
		GeneratedAtEpochMs: 100,
		Instances: map[string]fleetSnapshotRecord{
			"a": {
				RebootStatus:     "Scheduled",
				StatusChangedUTC: "2026-01-01 00:00:00",
			},
		},
	}
	if err := saveFleetSnapshot(path, in); err != nil {
		t.Fatalf("saveFleetSnapshot error: %v", err)
	}
	out, err := loadFleetSnapshot(path)
	if err != nil {
		t.Fatalf("loadFleetSnapshot error: %v", err)
	}
	if out.Instances["a"].RebootStatus != "Scheduled" {
		t.Fatalf("unexpected snapshot payload: %+v", out.Instances["a"])
	}
}
