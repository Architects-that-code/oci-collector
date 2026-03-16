package compute

import (
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"
)

func TestToJSONProducesValidPayload(t *testing.T) {
	summaries := []InstanceMetricsSummary{
		{
			Region:      "us-phoenix-1",
			Compartment: "root",
			InstanceID:  "ocid1.instance.oc1..abc",
			Name:        "worker-1",
			CpuDay:      InstanceMetrics{Avg: 1.5, P95: 2.0, P99: 2.2, Max: 3.0, SampleCount: 10},
		},
	}

	payload, err := ToJSON(summaries)
	if err != nil {
		t.Fatalf("ToJSON error: %v", err)
	}

	var got []InstanceMetricsSummary
	if err := json.Unmarshal([]byte(payload), &got); err != nil {
		t.Fatalf("invalid json payload: %v", err)
	}
	if len(got) != 1 || got[0].Name != "worker-1" {
		t.Fatalf("unexpected decoded payload: %+v", got)
	}
}

func TestToCSVIncludesHeaderAndNoteAggregation(t *testing.T) {
	summaries := []InstanceMetricsSummary{
		{
			Region:         "us-phoenix-1",
			Compartment:    "root",
			AvailabilityAD: "Uocm:PHX-AD-1",
			FaultDomain:    "FAULT-DOMAIN-1",
			InstanceID:     "ocid1.instance.oc1..abc",
			Name:           "worker-1",
			Shape:          "VM.Standard.E4.Flex",
			AgentInstalled: true,
			CpuDay:         InstanceMetrics{Avg: 1.00, P95: 2.00, P99: 3.00, Max: 4.00, SampleCount: 5, Note: "missing data"},
			CpuWeek:        InstanceMetrics{Avg: 1.10, P95: 2.10, P99: 3.10, Max: 4.10, SampleCount: 6},
			CpuMonth:       InstanceMetrics{Avg: 1.20, P95: 2.20, P99: 3.20, Max: 4.20, SampleCount: 7, Note: "sparse"},
			Note:           "base",
		},
	}

	csvText := ToCSV(summaries)
	r := csv.NewReader(strings.NewReader(csvText))
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("csv parse error: %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("expected header + row, got %d records", len(records))
	}
	header := records[0]
	if header[0] != "region" || header[23] != "note" {
		t.Fatalf("unexpected header: %v", header)
	}
	row := records[1]
	if row[0] != "us-phoenix-1" || row[5] != "worker-1" || row[7] != "true" {
		t.Fatalf("unexpected fixed columns: %v", row)
	}
	if row[8] != "1.00" || row[18] != "1.20" {
		t.Fatalf("unexpected metric formatting: %v", row)
	}
	if !strings.Contains(row[23], "base") || !strings.Contains(row[23], "cpu_day:missing data") || !strings.Contains(row[23], "cpu_month:sparse") {
		t.Fatalf("unexpected note column: %q", row[23])
	}
}
