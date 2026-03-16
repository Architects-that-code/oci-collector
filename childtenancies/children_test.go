package children

import (
	"strings"
	"testing"
)

func TestLifeCycleStats(t *testing.T) {
	items := []TenancyCollector{
		{LifecycleState: "ACTIVE"},
		{LifecycleState: "ACTIVE"},
		{LifecycleState: "INACTIVE"},
	}

	s := lifeCycleStats(items)
	if !strings.Contains(s, "Lifecycle State Stats:") {
		t.Fatalf("missing header: %q", s)
	}
	if !strings.Contains(s, "2 ACTIVE") {
		t.Fatalf("missing ACTIVE count in %q", s)
	}
	if !strings.Contains(s, "1 INACTIVE") {
		t.Fatalf("missing INACTIVE count in %q", s)
	}
}
