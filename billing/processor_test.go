package billing

import (
	"testing"
	"time"
)

func TestExtractDateFromPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		want    time.Time
		wantErr bool
	}{
		{
			name: "valid nested date",
			path: "/tmp/reports/2026/03/16/file.csv.gz",
			want: time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC),
		},
		{
			name:    "missing date",
			path:    "/tmp/reports/no-date/file.csv.gz",
			wantErr: true,
		},
		{
			name:    "invalid date pattern",
			path:    "/tmp/reports/2026/99/99/file.csv.gz",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := extractDateFromPath(tc.path)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for path %q", tc.path)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !got.Equal(tc.want) {
				t.Fatalf("unexpected date: got %v, want %v", got, tc.want)
			}
		})
	}
}
