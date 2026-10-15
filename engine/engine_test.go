package engine

import (
	"testing"

	"github.com/23dd/animation/scene"
	"github.com/23dd/animation/timeline"
)

func TestPlayPauseResumeAndReplay(t *testing.T) {
	tl := timeline.New()
	tl.Wait(2)
	clock := &FixedClock{}
	e := New(scene.New(320, 180), tl, nil)
	e.Clock = clock

	clock.Current = .5
	e.Update()
	e.Pause()
	clock.Current = 1
	e.Update()
	if state := e.State(); state.Playing || state.Time != .5 {
		t.Fatalf("paused state = %+v", state)
	}

	e.Play()
	e.Update()
	if state := e.State(); !state.Playing || state.Time != 1 {
		t.Fatalf("resumed state = %+v", state)
	}

	clock.Current = 3
	e.Update()
	if state := e.State(); state.Playing || state.Time != 2 {
		t.Fatalf("completed state = %+v", state)
	}
	e.Play()
	if state := e.State(); !state.Playing || state.Time != 0 {
		t.Fatalf("replay state = %+v", state)
	}
}
