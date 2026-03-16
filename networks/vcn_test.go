package network

import (
	"encoding/json"
	"testing"
)

func TestVcnCollectorJSONTags(t *testing.T) {
	v := VcnCollector{Region: "us-phoenix-1", CompartmentName: "root"}
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if _, ok := m["region"]; !ok {
		t.Fatalf("missing region key in json: %v", m)
	}
	if _, ok := m["compartmentname"]; !ok {
		t.Fatalf("missing compartmentname key in json: %v", m)
	}
	if _, ok := m["vcn"]; !ok {
		t.Fatalf("missing vcn key in json: %v", m)
	}
	if _, ok := m["subnets"]; !ok {
		t.Fatalf("missing subnets key in json: %v", m)
	}
}
