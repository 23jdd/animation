package timeline_test

import (
	"math"
	"testing"

	"github.com/23dd/animation/animation"
	"github.com/23dd/animation/geometry"
	"github.com/23dd/animation/mobject"
	"github.com/23dd/animation/timeline"
)

func TestSeekIsDeterministic(t *testing.T) {
	c := mobject.NewCircle(10)
	c.SetPosition(geometry.V2(10, 0))
	tl := timeline.New()
	tl.Add(0, animation.MoveTo(c, geometry.V2(110, 0)))
	tl.Add(1, animation.MoveTo(c, geometry.V2(210, 0)))
	tl.Seek(1.5)
	if math.Abs(c.Position().X-160) > .001 {
		t.Fatalf("at 1.5 got %v", c.Position().X)
	}
	tl.Seek(.5)
	if math.Abs(c.Position().X-60) > .001 {
		t.Fatalf("backward seek got %v", c.Position().X)
	}
	tl.Seek(1.5)
	if math.Abs(c.Position().X-160) > .001 {
		t.Fatalf("repeated seek got %v", c.Position().X)
	}
}

func TestSequenceAndParallelDurations(t *testing.T) {
	c := mobject.NewCircle(1)
	a := animation.Sequence(animation.MoveBy(c, geometry.V2(1, 0)), animation.MoveBy(c, geometry.V2(1, 0)))
	if a.Duration() != 2 {
		t.Fatalf("sequence duration %v", a.Duration())
	}
	p := animation.Parallel(a, animation.WithTiming(animation.FadeOut(c), 3, animation.Linear))
	if p.Duration() != 3 {
		t.Fatalf("parallel duration %v", p.Duration())
	}
}
