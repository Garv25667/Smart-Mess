// Package clock provides a Clock interface so tests can fix "now".
package clock

import "time"

// Clock provides the current time. Use Real() in production and
// Fixed() in tests.
type Clock interface {
	Now() time.Time
}

// realClock reads the actual system clock.
type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

// Real returns a Clock that uses the system time.
func Real() Clock { return realClock{} }

// fixedClock always returns the same instant.
type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time { return c.t }

// Fixed returns a Clock frozen at t.
func Fixed(t time.Time) Clock { return fixedClock{t: t} }
