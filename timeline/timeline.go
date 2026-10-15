package timeline

import (
	"github.com/23dd/animation/animation"
	"math"
	"sort"
	"sync"
)

type Timeline struct {
	mu           sync.Mutex
	Tracks       []Track
	Time, Length float64
}

func New() *Timeline { return &Timeline{} }
func (t *Timeline) Add(start float64, a animation.Animation) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Tracks = append(t.Tracks, Track{Start: start, Animation: a})
	sort.SliceStable(t.Tracks, func(i, j int) bool { return t.Tracks[i].Start < t.Tracks[j].Start })
	t.Length = math.Max(t.Length, start+a.Duration())
}
func (t *Timeline) Append(a animation.Animation) float64 {
	t.mu.Lock()
	start := t.Length
	t.mu.Unlock()
	t.Add(start, a)
	return start
}
func (t *Timeline) Wait(seconds float64) {
	if seconds < 0 {
		return
	}
	t.mu.Lock()
	t.Length += seconds
	t.mu.Unlock()
}
func (t *Timeline) Duration() float64 { t.mu.Lock(); defer t.mu.Unlock(); return t.Length }
func (t *Timeline) Seek(at float64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	at = math.Max(0, math.Min(t.Length, at))
	for i := len(t.Tracks) - 1; i >= 0; i-- {
		t.Tracks[i].Animation.Reset()
	}
	for _, track := range t.Tracks {
		if at < track.Start {
			continue
		}
		a := track.Animation
		a.Start()
		d := a.Duration()
		if d == 0 || at >= track.Start+d {
			a.Finish()
		} else {
			a.Update((at - track.Start) / d)
		}
	}
	t.Time = at
}
func (t *Timeline) Reset() { t.Seek(0) }
