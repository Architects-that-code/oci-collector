package autonomous

import "testing"

func TestSafeStr(t *testing.T) {
	if got := safeStr(nil); got != "" {
		t.Fatalf("safeStr(nil)=%q, want empty", got)
	}
	s := "adb-name"
	if got := safeStr(&s); got != "adb-name" {
		t.Fatalf("safeStr(&s)=%q, want adb-name", got)
	}
}
