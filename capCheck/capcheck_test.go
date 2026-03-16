package capacity

import (
	"reflect"
	"testing"
)

func TestMakeInstanceShape(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{in: "E4", want: "VM.Standard.E4.Flex"},
		{in: "E6", want: "VM.Standard.E6.Flex"},
		{in: "A1", want: "VM.Standard.A1.Flex"},
		{in: "X9", want: "VM.Standard3.Flex"},
		{in: "UNKNOWN", want: ""},
	}

	for _, tc := range tests {
		if got := makeInstanceShape(tc.in); got != tc.want {
			t.Fatalf("makeInstanceShape(%q)=%q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestMakeInstanceFamilyShapes(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{in: "AMD", want: []string{"VM.Standard.E4.Flex", "VM.Standard.E3.Flex", "VM.Standard.E5.Flex", "VM.Standard.E6.Flex"}},
		{in: "INTEL", want: []string{"VM.Standard3.Flex"}},
		{in: "ARM", want: []string{"VM.Standard.A1.Flex", "VM.Standard.A2.Flex", "VM.Standard.A4.Flex"}},
		{in: "OTHER", want: nil},
	}

	for _, tc := range tests {
		if got := makeInstanceFAMILYShapes(tc.in); !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("makeInstanceFAMILYShapes(%q)=%v, want %v", tc.in, got, tc.want)
		}
	}
}
