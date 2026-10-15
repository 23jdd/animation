package animation

import (
	"github.com/23dd/animation/geometry"
	"github.com/23dd/animation/mobject"
)

// DrawLine draws a line progressively from its start to its end. When tip is
// non-nil, that object follows the drawing head and is hidden on completion.
func DrawLine(target *mobject.Line, tip mobject.Mobject) Animation {
	var from float64
	var tipVisible bool
	var tipPos geometry.Vec2
	return New(1, Linear, func(float64) {
		from = target.Base().DrawProgress
		target.Base().IsVisible = true
		if tip != nil {
			tipVisible = tip.Base().IsVisible
			tipPos = tip.Base().Transform.Position
			tip.Base().IsVisible = true
		}
	}, func(t float64) {
		target.Base().DrawProgress = t
		if tip != nil {
			head := target.Base().Transform.Matrix().Apply(target.Start.Lerp(target.End, t))
			tip.Base().Transform.Position = head
		}
	}, func(float64) {
		if tip != nil {
			tip.Base().IsVisible = false
		}
	}, func(float64) {
		target.Base().DrawProgress = from
		if tip != nil {
			tip.Base().Transform.Position = tipPos
			tip.Base().IsVisible = tipVisible
		}
	})
}

// DrawLineFromCenter reveals a line by growing it outward from its midpoint.
func DrawLineFromCenter(target *mobject.Line) Animation {
	var start, end geometry.Vec2
	var fromProgress float64
	return New(1, Linear, func(float64) {
		start = target.Start
		end = target.End
		fromProgress = target.Base().DrawProgress
		target.Base().DrawProgress = 1
		target.Base().IsVisible = true
	}, func(t float64) {
		mid := start.Lerp(end, 0.5)
		target.Start = mid.Lerp(start, t)
		target.End = mid.Lerp(end, t)
	}, nil, func(float64) {
		target.Start = start
		target.End = end
		target.Base().DrawProgress = fromProgress
	})
}
