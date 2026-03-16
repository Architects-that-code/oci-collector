package limits

import (
	"encoding/json"
	"testing"
)

func TestLimitsCollectorJSONSerialization(t *testing.T) {
	data := []LimitsCollector{
		{
			Region:    "us-phoenix-1",
			Service:   "compute",
			Limitname: "vm-standard-e4",
			Avail:     500,
			Used:      22,
		},
	}

	b, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var out []LimitsCollector
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if len(out) != 1 {
		t.Fatalf("expected 1 item, got %d", len(out))
	}
	if out[0].Region != "us-phoenix-1" || out[0].Service != "compute" || out[0].Limitname != "vm-standard-e4" {
		t.Fatalf("unexpected payload after round-trip: %+v", out[0])
	}
}
