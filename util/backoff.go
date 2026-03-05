package util

import (
	crand "crypto/rand"
	"encoding/binary"
	mathrand "math/rand"
	"strings"
	"time"
)

const (
	maxBackoffDuration = 30 * time.Second
	maxRetryAttempts   = 8
)

func init() {
	// Seed math/rand once so jitter isn’t deterministic across runs.
	// Use crypto/rand to avoid depending on wall-clock time.
	var b [8]byte
	if _, err := crand.Read(b[:]); err == nil {
		mathrand.Seed(int64(binary.LittleEndian.Uint64(b[:])))
	} else {
		mathrand.Seed(time.Now().UnixNano())
	}
}

// BackoffWithJitter returns an exponential backoff duration capped at maxBackoffDuration with jitter.
func BackoffWithJitter(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	base := time.Second
	duration := base << attempt
	if duration > maxBackoffDuration {
		duration = maxBackoffDuration
	}
	jitter := time.Duration(mathrand.Intn(300)) * time.Millisecond
	return duration + jitter
}

// IsTooManyRequests detects throttling errors based on the error message.
func IsTooManyRequests(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	if strings.Contains(msg, "TooManyRequests") {
		return true
	}
	if strings.Contains(msg, "429") {
		return true
	}
	if strings.Contains(strings.ToLower(msg), "throttl") {
		return true
	}
	return false
}

// RetryWithBackoff retries the given operation when throttled and returns the final response or error.
func RetryWithBackoff[T any](operation func() (T, error)) (T, error) {
	var result T
	var err error
	for attempt := 0; attempt < maxRetryAttempts; attempt++ {
		result, err = operation()
		if err == nil {
			return result, nil
		}
		if !IsTooManyRequests(err) {
			return result, err
		}
		time.Sleep(BackoffWithJitter(attempt))
	}
	return result, err
}
