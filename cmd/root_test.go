package cmd

import (
	"bytes"
	"sort"
	"strings"
	"testing"
)

func executeRoot(t *testing.T, args ...string) string {
	t.Helper()
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("root execute error: %v", err)
	}
	return buf.String()
}

func TestRootRegistersExpectedCommands(t *testing.T) {
	got := make([]string, 0, len(rootCmd.Commands()))
	for _, c := range rootCmd.Commands() {
		got = append(got, c.Name())
	}
	sort.Strings(got)

	want := []string{
		"autonomous", "billing", "capability", "capacity", "children", "cloudadvisor", "compute", "config",
		"groups", "limits", "network", "object", "peeps", "policies", "schedule", "search", "support",
	}
	sort.Strings(want)

	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("unexpected commands\n got: %v\nwant: %v", got, want)
	}
}

func TestRootShowsHelpByDefault(t *testing.T) {
	out := executeRoot(t)
	if !strings.Contains(out, "A loose collection of tools to help manage and monitor your OCI tenancy.") {
		t.Fatalf("expected root help output, got: %q", out)
	}
}

func TestSearchWithoutQueryShowsHelp(t *testing.T) {
	out := executeRoot(t, "search")
	if !strings.Contains(out, "search string") {
		t.Fatalf("expected search help output, got: %q", out)
	}
}

func TestLimitsWithoutRunShowsHelp(t *testing.T) {
	out := executeRoot(t, "limits")
	if !strings.Contains(out, "fetch limits in all regions") {
		t.Fatalf("expected limits help output, got: %q", out)
	}
}

func TestNetworkWithoutRunShowsHelp(t *testing.T) {
	out := executeRoot(t, "network")
	if !strings.Contains(out, "fetch all vcn in all regions") {
		t.Fatalf("expected network help output, got: %q", out)
	}
}

func TestScheduleWithoutRunShowsHelp(t *testing.T) {
	out := executeRoot(t, "schedule")
	if !strings.Contains(out, "fetch schedule") {
		t.Fatalf("expected schedule help output, got: %q", out)
	}
}

func TestSupportHelpShowsListFlag(t *testing.T) {
	out := executeRoot(t, "support", "--help")
	if !strings.Contains(out, "--list") {
		t.Fatalf("expected support help to include --list flag, got: %q", out)
	}
}

func TestNetworkHelpShowsRPCFlag(t *testing.T) {
	out := executeRoot(t, "network", "--help")
	if !strings.Contains(out, "--rpc") {
		t.Fatalf("expected network help to include --rpc flag, got: %q", out)
	}
}

func TestSearchHelpShowsSearchStringFlag(t *testing.T) {
	out := executeRoot(t, "search", "--help")
	if !strings.Contains(out, "--searchstring") {
		t.Fatalf("expected search help to include --searchstring flag, got: %q", out)
	}
}

func TestPeepsWithoutRunShowsHelp(t *testing.T) {
	out := executeRoot(t, "peeps")
	if !strings.Contains(out, "fetch users") {
		t.Fatalf("expected peeps help output, got: %q", out)
	}
}

func TestGroupsWithoutRunShowsHelp(t *testing.T) {
	out := executeRoot(t, "groups")
	if !strings.Contains(out, "fetch groups") {
		t.Fatalf("expected groups help output, got: %q", out)
	}
}

func TestPoliciesWithoutRunShowsHelp(t *testing.T) {
	out := executeRoot(t, "policies")
	if !strings.Contains(out, "fetch policy") {
		t.Fatalf("expected policies help output, got: %q", out)
	}
}

func TestObjectWithoutRunShowsHelp(t *testing.T) {
	out := executeRoot(t, "object")
	if !strings.Contains(out, "fetch object storage") {
		t.Fatalf("expected object help output, got: %q", out)
	}
}

func TestChildrenWithoutRunShowsHelp(t *testing.T) {
	out := executeRoot(t, "children")
	if !strings.Contains(out, "fetch child tenancies") {
		t.Fatalf("expected children help output, got: %q", out)
	}
}

func TestAutonomousWithoutRunShowsHelp(t *testing.T) {
	out := executeRoot(t, "autonomous")
	if !strings.Contains(out, "list autonomous databases") {
		t.Fatalf("expected autonomous help output, got: %q", out)
	}
}

func TestSupportWithoutListShowsHelp(t *testing.T) {
	out := executeRoot(t, "support")
	if !strings.Contains(out, "--list") {
		t.Fatalf("expected support help output requiring --list, got: %q", out)
	}
}

func TestBillingWithoutActionShowsHelp(t *testing.T) {
	out := executeRoot(t, "billing")
	if !strings.Contains(out, "--download") {
		t.Fatalf("expected billing help output requiring action flags, got: %q", out)
	}
}

func TestRootHelpDoesNotShowToggleFlag(t *testing.T) {
	out := executeRoot(t, "--help")
	if strings.Contains(out, "--toggle") {
		t.Fatalf("did not expect --toggle in root help, got: %q", out)
	}
}

func TestRootHelpShowsProfileOverrideFlag(t *testing.T) {
	out := executeRoot(t, "--help")
	if !strings.Contains(out, "--profile") {
		t.Fatalf("expected root help to include --profile flag, got: %q", out)
	}
}

func TestCapacityHelpDoesNotShowRemovedAdFdFlags(t *testing.T) {
	out := executeRoot(t, "capacity", "--help")
	if strings.Contains(out, "--ad") || strings.Contains(out, "--fd") {
		t.Fatalf("did not expect --ad/--fd in capacity help, got: %q", out)
	}
}

func TestComputeHelpDoesNotShowMetricsTypesFlag(t *testing.T) {
	out := executeRoot(t, "compute", "--help")
	if strings.Contains(out, "--metrics-types") {
		t.Fatalf("did not expect --metrics-types in compute help, got: %q", out)
	}
}

func TestComputeHelpShowsFleetJSONFormat(t *testing.T) {
	out := executeRoot(t, "compute", "--help")
	if !strings.Contains(out, "fleet-json") {
		t.Fatalf("expected compute help to include fleet-json format, got: %q", out)
	}
}
