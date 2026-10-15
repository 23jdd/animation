package engine

import (
	"github.com/23dd/animation/renderer"
	"github.com/23dd/animation/scene"
	"github.com/23dd/animation/timeline"
	"image"
	"sync"
)

type Engine struct {
	mu          sync.RWMutex
	Scene       *scene.Scene
	Timeline    *timeline.Timeline
	Renderer    renderer.Renderer
	Clock       Clock
	Playing     bool
	Speed       float64
	last        float64
	ScriptError error
}

func New(s *scene.Scene, t *timeline.Timeline, r renderer.Renderer) *Engine {
	return &Engine{Scene: s, Timeline: t, Renderer: r, Clock: NewRealtimeClock(), Playing: true, Speed: 1}
}
func (e *Engine) Update() {
	e.mu.Lock()
	defer e.mu.Unlock()
	now := e.Clock.Time()
	if !e.Playing {
		now = e.last
	}
	if duration := e.Timeline.Duration(); duration > 0 && now >= duration {
		now = duration
		if e.Playing {
			e.Playing = false
			e.Clock.Pause()
		}
	}
	e.Timeline.Seek(now)
	e.Scene.Update(now, now-e.last)
	e.last = now
}
func (e *Engine) Render() image.Image {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.Renderer.Render(e.Scene)
}
func (e *Engine) Play() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if duration := e.Timeline.Duration(); duration > 0 && e.last >= duration {
		e.Clock.Reset()
		e.last = 0
		e.Timeline.Seek(0)
	}
	e.Playing = true
	e.Clock.Resume()
}
func (e *Engine) Pause() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Playing = false
	e.Clock.Pause()
}
func (e *Engine) Toggle() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.Playing {
		e.Playing = false
		e.Clock.Pause()
	} else {
		if duration := e.Timeline.Duration(); duration > 0 && e.last >= duration {
			e.Clock.Reset()
			e.last = 0
			e.Timeline.Seek(0)
		}
		e.Playing = true
		e.Clock.Resume()
	}
}
func (e *Engine) Restart() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Clock.Reset()
	if !e.Playing {
		e.Clock.Pause()
	}
	e.last = 0
	e.Timeline.Seek(0)
}
func (e *Engine) Seek(v float64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if v < 0 {
		v = 0
	}
	if d := e.Timeline.Duration(); v > d {
		v = d
	}
	if c, ok := e.Clock.(interface{ SetTime(float64) }); ok {
		c.SetTime(v)
	}
	e.Timeline.Seek(v)
	e.last = e.Timeline.Time
}
func (e *Engine) SetSpeed(v float64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if v > 0 {
		e.Speed = v
		e.Clock.SetSpeed(v)
	}
}
func (e *Engine) Swap(s *scene.Scene, t *timeline.Timeline, err error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if err != nil {
		e.ScriptError = err
		return
	}
	at := e.last
	e.Scene = s
	e.Timeline = t
	t.Seek(at)
	e.ScriptError = nil
}
