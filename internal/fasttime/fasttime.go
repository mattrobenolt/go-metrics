// See: https://go.withmatt.com/fasttime
package fasttime

import (
	"context"
	"sync/atomic"
	"time"
)

var monotonicRoot = time.Now()

// Instant is roughly equivalent to a [time.Time], but backed by a nanosecond
// duration since process start.
type Instant int64

// ToTime converts an [Instant] into a [time.Time].
func (i Instant) ToTime() time.Time {
	return monotonicRoot.Add(time.Duration(i))
}

// Sub subtracts two [Instant] s, similar to [time.Time.Sub].
func (i Instant) Sub(u Instant) time.Duration {
	return time.Duration(i - u)
}

// String formats the [Instant] using the underlying [time.Time.String].
func (i Instant) String() string {
	return i.ToTime().String()
}

// Now is roughly equivalent to [time.Now], but returns an [Instant].
func Now() Instant {
	return Instant(time.Since(monotonicRoot))
}

// Since is roughly equivalent to [time.Since], but operates on [Instant] s.
func Since(i Instant) time.Duration {
	return Now().Sub(i)
}

// Clock is an instance of a clock whose time increments roughly at a
// configured granularity, but lookups are effectively free relative to
// normal [time.Now].
type Clock struct {
	ctx    context.Context
	cancel context.CancelFunc
	now    atomic.Int64
}

// NewClock creates a new [Clock] configured to tick at approximately
// granularity intervals. [Clock] is running when created, and may be stopped
// by calling [Clock.Stop]. A stopped [Clock] cannot be resumed.
func NewClock(granularity time.Duration) *Clock {
	ctx, cancel := context.WithCancel(context.Background())
	c := &Clock{ctx: ctx, cancel: cancel}
	c.now.Store(int64(Now()))
	go c.run(granularity)
	return c
}

// Now returns an [Instant] that represents the current cached time.
// The [Instant] returned will never be in the future, but will always be
// less than or equal to the actual current time.
func (c *Clock) Now() Instant {
	return Instant(c.now.Load())
}

// Since returns [time.Duration] since the [Instant] relative to the [Clock]'s
// current time.
func (c *Clock) Since(i Instant) time.Duration {
	return c.Now().Sub(i)
}

// Stop stops the [Clock] ticker and cannot be resumed.
func (c *Clock) Stop() {
	c.cancel()
}

func (c *Clock) run(granularity time.Duration) {
	t := time.NewTicker(granularity)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			c.now.Store(int64(Now()))
		case <-c.ctx.Done():
			return
		}
	}
}
