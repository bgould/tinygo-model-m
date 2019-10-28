package timer

import "time"

type Timer struct {
	start    int64
	interval int64
}

func Wait(duration time.Duration) {
	New(duration).WaitUntilExpired()
}

func New(interval time.Duration) Timer {
	return Timer{
		start:    time.Now().UnixNano(),
		interval: int64(interval),
	}
}

func (t Timer) Expired() bool {
	return time.Now().UnixNano() > (t.start + t.interval)
}

func (t Timer) WaitUntilExpired() {
	for !t.Expired() {
	}
}
