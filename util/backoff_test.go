package util

import (
	"errors"
	"testing"
	"time"
)

func TestBackoffWithJitterBounds(t *testing.T) {
	tests := []struct {
		name       string
		attempt    int
		min, max   time.Duration
	}{
		{name: "negative attempt clamps to zero", attempt: -1, min: time.Second, max: time.Second + 300*time.Millisecond},
		{name: "attempt zero", attempt: 0, min: time.Second, max: time.Second + 300*time.Millisecond},
		{name: "attempt one", attempt: 1, min: 2 * time.Second, max: 2*time.Second + 300*time.Millisecond},
		{name: "large attempt is capped", attempt: 99, min: 30 * time.Second, max: 30*time.Second + 300*time.Millisecond},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := BackoffWithJitter(tc.attempt)
			if d < tc.min || d > tc.max {
				t.Fatalf("unexpected backoff: got %v, want between %v and %v", d, tc.min, tc.max)
			}
		})
	}
}

func TestIsTooManyRequests(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "too many requests token", err: errors.New("service error: TooManyRequests"), want: true},
		{name: "http 429", err: errors.New("status=429 from upstream"), want: true},
		{name: "throttling text", err: errors.New("request throttled by service"), want: true},
		{name: "non throttle", err: errors.New("permission denied"), want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := IsTooManyRequests(tc.err)
			if got != tc.want {
				t.Fatalf("IsTooManyRequests()=%v, want %v", got, tc.want)
			}
		})
	}
}

func TestRetryWithBackoffReturnsImmediatelyOnSuccess(t *testing.T) {
	calls := 0
	result, err := RetryWithBackoff(func() (int, error) {
		calls++
		return 42, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != 42 {
		t.Fatalf("unexpected result: got %d, want 42", result)
	}
	if calls != 1 {
		t.Fatalf("expected one call, got %d", calls)
	}
}

func TestRetryWithBackoffReturnsNonThrottleErrorWithoutRetry(t *testing.T) {
	calls := 0
	_, err := RetryWithBackoff(func() (int, error) {
		calls++
		return 0, errors.New("boom")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if calls != 1 {
		t.Fatalf("expected one call, got %d", calls)
	}
}

func TestRetryWithBackoffRetriesThrottleThenSucceeds(t *testing.T) {
	calls := 0
	result, err := RetryWithBackoff(func() (string, error) {
		calls++
		if calls == 1 {
			return "", errors.New("429 throttled")
		}
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "ok" {
		t.Fatalf("unexpected result: got %q, want %q", result, "ok")
	}
	if calls != 2 {
		t.Fatalf("expected two calls, got %d", calls)
	}
}
