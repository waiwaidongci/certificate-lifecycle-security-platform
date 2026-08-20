package clock

import "time"

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (RealClock) Now() time.Time {
	return time.Now().UTC()
}

type FrozenClock struct {
	At time.Time
}

func (f FrozenClock) Now() time.Time {
	return f.At
}
