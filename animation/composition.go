package animation

import "math"

type composition struct {
	children []Animation
	starts   []float64
	duration float64
}

func (c *composition) Duration() float64 { return c.duration }
func (c *composition) Start()            {}
func (c *composition) Update(t float64)  { seekChildren(c.children, c.starts, c.duration*clamp(t)) }
func (c *composition) Finish()           { c.Update(1) }
func (c *composition) Reset() {
	for i := len(c.children) - 1; i >= 0; i-- {
		c.children[i].Reset()
	}
}

func Sequence(items ...Animation) Animation {
	c := &composition{children: items, starts: make([]float64, len(items))}
	for i, a := range items {
		c.starts[i] = c.duration
		c.duration += a.Duration()
	}
	return c
}
func Parallel(items ...Animation) Animation {
	c := &composition{children: items, starts: make([]float64, len(items))}
	for _, a := range items {
		c.duration = math.Max(c.duration, a.Duration())
	}
	return c
}
func Stagger(delay float64, items ...Animation) Animation {
	c := &composition{children: items, starts: make([]float64, len(items))}
	for i, a := range items {
		c.starts[i] = float64(i) * delay
		c.duration = math.Max(c.duration, c.starts[i]+a.Duration())
	}
	return c
}
func seekChildren(items []Animation, starts []float64, time float64) {
	for i, a := range items {
		if time < starts[i] {
			continue
		}
		a.Start()
		d := a.Duration()
		if d == 0 || time >= starts[i]+d {
			a.Finish()
		} else {
			a.Update((time - starts[i]) / d)
		}
	}
}

type repeated struct {
	inner Animation
	count int
	yoyo  bool
}

func Repeat(a Animation, count int) Animation {
	if count < 0 {
		count = 0
	}
	return &repeated{inner: a, count: count}
}
func Yoyo(a Animation, count int) Animation {
	if count < 0 {
		count = 0
	}
	return &repeated{inner: a, count: count, yoyo: true}
}
func (r *repeated) Duration() float64 { return r.inner.Duration() * float64(r.count) }
func (r *repeated) Start()            {}
func (r *repeated) Update(t float64) {
	if r.count == 0 {
		return
	}
	x := clamp(t) * float64(r.count)
	cycle := int(math.Min(float64(r.count-1), math.Floor(x)))
	local := x - float64(cycle)
	r.inner.Reset()
	r.inner.Start()
	if r.yoyo && cycle%2 == 1 {
		local = 1 - local
	}
	r.inner.Update(local)
}
func (r *repeated) Finish() { r.Update(1) }
func (r *repeated) Reset()  { r.inner.Reset() }
