package metrics

import (
	"io"
	"sync"
	"testing"
	"time"

	"go.withmatt.com/metrics/internal/fasttime"
)

// Check for a race between the prometheus writer (which deletes expired sets)
// and an Inc().
func TestSetTTLRace(t *testing.T) {
	set := NewSet()
	sv := set.NewSetVecWithTTL("set_group_label", 7*24*time.Hour)
	sv.SetIsActive(func(s *Set) bool {
		active, _ := s.GetMetricUint64("counter")
		return active > 0
	})
	metric := sv.NewUint64Vec("counter")
	child := sv.WithLabelValue("a")

	// Model a set that was idle for longer than the production TTL. This
	// happens BEFORE the race; neither goroutine needs to pause for 7 days.
	child.lastUsed.Store(fastClock().Now() - fasttime.Instant(child.ttl+time.Second))

	expirationDecided := make(chan struct{})
	resumeDeletion := make(chan struct{})
	oldHook := testHookBeforeSetDelete
	testHookBeforeSetDelete = func(s *Set) {
		if s == child {
			// RACER 1 has checked both activity and TTL, but has not
			// deleted the set. Model descheduling at that exact point.
			close(expirationDecided)
			<-resumeDeletion
		}
	}
	defer func() { testHookBeforeSetDelete = oldHook }()

	// RACER 1: a scrape from GET /metrics.
	scraped := make(chan error, 1)
	go func() {
		_, err := set.WritePrometheus(io.Discard)
		scraped <- err
	}()
	finishScrape := sync.OnceValue(func() error {
		close(resumeDeletion)
		return <-scraped
	})
	// Release and join before restoring the hook, even on setup failures.
	defer func() {
		if err := finishScrape(); err != nil {
			t.Errorf("WritePrometheus: %v", err)
		}
	}()

	select {
	case <-expirationDecided:
	case <-time.After(5 * time.Second):
		t.Fatal("precondition failed: scraper did not reach deletion of the expired set")
	}

	// RACER 2: calls Inc() to count a new connection, then later calls
	// Dec() when the connection closes.

	// The WithLabelValues call here races with the metrics scrape. This is
	// the only call that needs to happen while RACER 1 is descheduled.
	m1 := metric.WithLabelValues("a")

	// The race condition has now occurred.
	// + RACER 1 will delete the set.
	// + RACER 2 will Inc() the old counter and Dec() the new one.

	if err := finishScrape(); err != nil {
		t.Fatalf("precondition failed: WritePrometheus: %v", err)
	}

	m1.Inc()
	if v := m1.Get(); v != 1 {
		t.Errorf("precondition failed: want counter to be 1, got %d", v)
	}

	// Later in RACER 2, the decrement looks up the counter again.
	m2 := metric.WithLabelValues("a")
	m2.Dec()
	if m1 != m2 {
		t.Errorf("race failed and an active counter was expired: Inc used %p, Dec used %p",
			m1, m2)
	}
	if v := m2.Get(); v != 0 {
		t.Errorf("after Inc,Dec, want counter to be 0 but got %d", v)
	}
}
