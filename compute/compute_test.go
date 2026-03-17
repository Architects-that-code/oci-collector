package compute

import (
	"encoding/csv"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/oracle/oci-go-sdk/core"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/identity"
)

func TestSafeStr(t *testing.T) {
	if got := safeStr(nil); got != "" {
		t.Fatalf("safeStr(nil)=%q, want empty string", got)
	}
	s := "value"
	if got := safeStr(&s); got != "value" {
		t.Fatalf("safeStr(&s)=%q, want %q", got, s)
	}
}

func TestIsTooManyRequests(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "TooManyRequests", err: errors.New("TooManyRequests from service"), want: true},
		{name: "429", err: errors.New("status code 429"), want: true},
		{name: "throttle", err: errors.New("throttled by server"), want: true},
		{name: "other", err: errors.New("unauthorized"), want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := isTooManyRequests(tc.err); got != tc.want {
				t.Fatalf("isTooManyRequests()=%v, want %v", got, tc.want)
			}
		})
	}
}

func TestBackoffWithJitterBounds(t *testing.T) {
	tests := []struct {
		attempt int
		min     time.Duration
		max     time.Duration
	}{
		{attempt: -1, min: time.Second, max: time.Second + 300*time.Millisecond},
		{attempt: 0, min: time.Second, max: time.Second + 300*time.Millisecond},
		{attempt: 2, min: 4 * time.Second, max: 4*time.Second + 300*time.Millisecond},
		{attempt: 42, min: 30 * time.Second, max: 30*time.Second + 300*time.Millisecond},
	}

	for _, tc := range tests {
		d := backoffWithJitter(tc.attempt)
		if d < tc.min || d > tc.max {
			t.Fatalf("attempt %d: got %v, want between %v and %v", tc.attempt, d, tc.min, tc.max)
		}
	}
}

func TestIsAgentInstalled(t *testing.T) {
	falseVal := false
	trueVal := true

	tests := []struct {
		name string
		inst core.Instance
		want bool
	}{
		{name: "no agent config", inst: core.Instance{}, want: false},
		{name: "monitoring disabled", inst: core.Instance{AgentConfig: &core.InstanceAgentConfig{IsMonitoringDisabled: &trueVal}}, want: false},
		{name: "monitoring enabled", inst: core.Instance{AgentConfig: &core.InstanceAgentConfig{IsMonitoringDisabled: &falseVal}}, want: true},
		{name: "config present and nil disabled pointer", inst: core.Instance{AgentConfig: &core.InstanceAgentConfig{}}, want: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := isAgentInstalled(tc.inst); got != tc.want {
				t.Fatalf("isAgentInstalled()=%v, want %v", got, tc.want)
			}
		})
	}
}

func TestGroupInventoryByRegionCompartment(t *testing.T) {
	compRoot := identity.Compartment{Name: common.String("root")}
	compApp := identity.Compartment{Name: common.String("app")}

	inventory := []InstanceInventory{
		{RegionName: "us-phoenix-1", Compartment: compRoot, Instance: core.Instance{Id: common.String("i1")}},
		{RegionName: "us-phoenix-1", Compartment: compRoot, Instance: core.Instance{Id: common.String("i2")}},
		{RegionName: "us-ashburn-1", Compartment: compApp, Instance: core.Instance{Id: common.String("i3")}},
	}

	groups := groupInventoryByRegionCompartment(inventory)
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}

	if groups[0].Region != "us-phoenix-1" || groups[0].Compartment != "root" || len(groups[0].Items) != 2 {
		t.Fatalf("unexpected first group: %+v", groups[0])
	}
	if groups[1].Region != "us-ashburn-1" || groups[1].Compartment != "app" || len(groups[1].Items) != 1 {
		t.Fatalf("unexpected second group: %+v", groups[1])
	}
}

func TestFlattenInstancesAndCSV(t *testing.T) {
	falseVal := false
	groups := []InstanceGroups{
		{
			Region:      "us-phoenix-1",
			Compartment: "root",
			Items: []InstanceInventory{
				{
					Instance: core.Instance{
						AvailabilityDomain: common.String("Uocm:PHX-AD-1"),
						FaultDomain:        common.String("FAULT-DOMAIN-1"),
						Id:                 common.String("ocid1.instance.oc1..abc"),
						DisplayName:        common.String("worker-1"),
						Shape:              common.String("VM.Standard.E4.Flex"),
						FreeformTags:       map[string]string{"env": "dev"},
						AgentConfig:        &core.InstanceAgentConfig{IsMonitoringDisabled: &falseVal},
					},
					AgentPlugins: []AgentPluginState{
						{Name: "Bastion", Status: "RUNNING"},
						{Name: "Vulnerability Scanning", Status: "STOPPED"},
					},
				},
			},
		},
	}

	rows := flattenInstances(groups)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	row := rows[0]
	if row.Region != "us-phoenix-1" || row.Compartment != "root" || row.DisplayName != "worker-1" {
		t.Fatalf("unexpected row payload: %+v", row)
	}
	if !row.AgentInstalled {
		t.Fatalf("expected agent installed=true, got false")
	}
	if got := formatAgentPlugins(row.AgentPlugins); got != "Bastion:RUNNING;Vulnerability Scanning:STOPPED" {
		t.Fatalf("unexpected formatted plugins: %q", got)
	}

	csvText := toCSV(rows)
	r := csv.NewReader(strings.NewReader(csvText))
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("csv parse error: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected header + 1 row, got %d records", len(records))
	}
	if records[0][0] != "region" || records[0][7] != "agentInstalled" || records[0][8] != "agentPlugins" {
		t.Fatalf("unexpected csv header: %v", records[0])
	}
	if records[1][0] != "us-phoenix-1" || records[1][5] != "worker-1" || records[1][7] != "true" || records[1][8] != "Bastion:RUNNING;Vulnerability Scanning:STOPPED" {
		t.Fatalf("unexpected csv row: %v", records[1])
	}
}

func TestIsAgentInstalledWithPlugins(t *testing.T) {
	if !isAgentInstalledWithPlugins(false, []AgentPluginState{{Name: "A", Status: "RUNNING"}}) {
		t.Fatalf("expected plugins to imply installed agent")
	}
	if isAgentInstalledWithPlugins(false, nil) {
		t.Fatalf("expected false when no plugins and config says not installed")
	}
	if !isAgentInstalledWithPlugins(true, nil) {
		t.Fatalf("expected true when config says installed")
	}
}

func TestFormatAgentPlugins(t *testing.T) {
	plugins := []AgentPluginState{
		{Name: "Bastion", Status: "RUNNING"},
		{Name: "Vulnerability Scanning", Status: "STOPPED"},
	}
	if got := formatAgentPlugins(plugins); got != "Bastion:RUNNING;Vulnerability Scanning:STOPPED" {
		t.Fatalf("unexpected plugin format: %q", got)
	}
	if got := formatAgentPlugins(nil); got != "" {
		t.Fatalf("expected empty plugin format for nil slice, got %q", got)
	}
}
