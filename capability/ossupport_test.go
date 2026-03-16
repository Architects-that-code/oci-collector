package capability

import "testing"

func TestMakeInstanceShape(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{in: "E4", want: "VM.Standard.E4.Flex"},
		{in: "E3", want: "VM.Standard.E3.Flex"},
		{in: "A1", want: "VM.Standard.A1.Flex"},
		{in: "X9", want: "VM.Standard3.Flex"},
		{in: "A2", want: ""},
	}
	for _, tc := range tests {
		if got := makeInstanceShape(tc.in); got != tc.want {
			t.Fatalf("makeInstanceShape(%q)=%q, want %q", tc.in, got, tc.want)
		}
	}
}
