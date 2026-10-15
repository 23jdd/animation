package engine

import "time"

type Clock interface {
	Time() float64
	Reset()
	Pause()
	Resume()
	SetSpeed(float64)
}
type RealtimeClock struct {
	started  time.Time
	pausedAt time.Time
	paused   time.Duration
	isPaused bool
	speed    float64
}

func NewRealtimeClock() *RealtimeClock { c := &RealtimeClock{speed: 1}; c.Reset(); return c }
func (c *RealtimeClock) Time() float64 {
	end := time.Now()
	if c.isPaused {
		end = c.pausedAt
	}
	return end.Sub(c.started.Add(c.paused)).Seconds() * c.speed
}
func (c *RealtimeClock) Reset() { c.started = time.Now(); c.paused = 0; c.isPaused = false }
func (c *RealtimeClock) Pause() {
	if !c.isPaused {
		c.pausedAt = time.Now()
		c.isPaused = true
	}
}
func (c *RealtimeClock) Resume() {
	if c.isPaused {
		c.paused += time.Since(c.pausedAt)
		c.isPaused = false
	}
}
func (c *RealtimeClock) SetSpeed(v float64) {
	if v > 0 {
		now := c.Time()
		c.started = time.Now().Add(-time.Duration(now / v * float64(time.Second)))
		c.paused = 0
		c.speed = v
		if c.isPaused {
			c.pausedAt = time.Now()
		}
	}
}
func (c *RealtimeClock) SetTime(v float64) {
	if v < 0 {
		v = 0
	}
	c.started = time.Now().Add(-time.Duration(v / c.speed * float64(time.Second)))
	c.paused = 0
	if c.isPaused {
		c.pausedAt = time.Now()
	}
}

type FixedClock struct{ Current, Step float64 }

func (c *FixedClock) Time() float64     { return c.Current }
func (c *FixedClock) Reset()            { c.Current = 0 }
func (c *FixedClock) Pause()            {}
func (c *FixedClock) Resume()           {}
func (c *FixedClock) SetSpeed(float64)  {}
func (c *FixedClock) Next()             { c.Current += c.Step }
func (c *FixedClock) SetTime(v float64) { c.Current = v }
