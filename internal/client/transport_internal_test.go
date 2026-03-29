package client

import (
	"testing"
	"time"
)

func TestComputeDelayJitter(t *testing.T) {
	base := 100 * time.Millisecond
	seen := make(map[time.Duration]bool)
	for i := 0; i < 100; i++ {
		d := computeDelay(base, 0)
		seen[d] = true
		// attempt 0: base * 1 = 100ms, jitter +/-20% = [80ms, 120ms]
		if d < 80*time.Millisecond || d > 120*time.Millisecond {
			t.Errorf("delay %v out of expected range [80ms, 120ms]", d)
		}
	}
	if len(seen) < 2 {
		t.Errorf("expected jitter to produce varied delays, got %d unique values", len(seen))
	}
}

func TestComputeDelayCap(t *testing.T) {
	base := time.Second
	d := computeDelay(base, 10) // 1s * 1024 = 1024s, capped at 30s + jitter
	maxWithJitter := 30*time.Second + 6*time.Second // 30s + 20%
	if d > maxWithJitter {
		t.Errorf("delay %v exceeds cap+jitter %v", d, maxWithJitter)
	}
	minWithJitter := 30*time.Second - 6*time.Second // 30s - 20%
	if d < minWithJitter {
		t.Errorf("delay %v below cap-jitter %v", d, minWithJitter)
	}
}
