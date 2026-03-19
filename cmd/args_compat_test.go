package cmd

import (
	"reflect"
	"testing"
)

func TestNormalizeProfileAliasArgs(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{
			name: "single dash profile becomes long option",
			in:   []string{"-profile", "PROD", "compute", "-r"},
			want: []string{"--profile", "PROD", "compute", "-r"},
		},
		{
			name: "single dash profile equals form becomes long option",
			in:   []string{"-profile=PROD", "compute", "-r"},
			want: []string{"--profile=PROD", "compute", "-r"},
		},
		{
			name: "double dash profile untouched",
			in:   []string{"--profile", "PROD", "compute", "-r"},
			want: []string{"--profile", "PROD", "compute", "-r"},
		},
		{
			name: "non profile args untouched",
			in:   []string{"billing", "-p", "reports", "-d"},
			want: []string{"billing", "-p", "reports", "-d"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeProfileAliasArgs(tc.in)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("normalizeProfileAliasArgs() got=%v want=%v", got, tc.want)
			}
		})
	}
}
