package animation

import "math"

type Animation interface {
	Duration() float64
	Start()
	Update(progress float64)
	Finish()
	Reset()
}

func clamp(t float64) float64 { return math.Max(0, math.Min(1, t)) }

type Func struct {
	duration                     float64
	ease                         Easing
	start, update, finish, reset func(float64)
	started                      bool
}

func New(duration float64, ease Easing, start, update, finish, reset func(float64)) *Func {
	if duration < 0 {
		duration = 0
	}
	if ease == nil {
		ease = Linear
	}
	return &Func{duration: duration, ease: ease, start: start, update: update, finish: finish, reset: reset}
}
func (a *Func) Duration() float64 { return a.duration }
func (a *Func) Start() {
	if !a.started {
		if a.start != nil {
			a.start(0)
		}
		a.started = true
	}
}
func (a *Func) Update(t float64) {
	if a.update != nil {
		a.update(a.ease(clamp(t)))
	}
}
func (a *Func) Finish() {
	a.Update(1)
	if a.finish != nil {
		a.finish(1)
	}
}
func (a *Func) Reset() {
	if a.started && a.reset != nil {
		a.reset(0)
	}
	a.started = false
}

// Timed changes the apparent duration and easing of another animation.
type Timed struct {
	Inner  Animation
	Length float64
	Ease   Easing
}

func WithTiming(a Animation, duration float64, ease Easing) Animation {
	if duration < 0 {
		duration = 0
	}
	if ease == nil {
		ease = Linear
	}
	return &Timed{Inner: a, Length: duration, Ease: ease}
}
func (a *Timed) Duration() float64 { return a.Length }
func (a *Timed) Start()            { a.Inner.Start() }
func (a *Timed) Update(t float64)  { a.Inner.Update(a.Ease(clamp(t))) }
func (a *Timed) Finish()           { a.Inner.Finish() }
func (a *Timed) Reset()            { a.Inner.Reset() }
