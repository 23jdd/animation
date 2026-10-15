package animation

import (
	"image/color"
	"math"

	"github.com/23dd/animation/geometry"
	"github.com/23dd/animation/mobject"
)

func lerpColor(a, b color.RGBA, t float64) color.RGBA {
	lerp := func(x, y uint8) uint8 { return uint8(math.Round(float64(x) + (float64(y)-float64(x))*t)) }
	return color.RGBA{lerp(a.R, b.R), lerp(a.G, b.G), lerp(a.B, b.B), lerp(a.A, b.A)}
}

// GrowFromCenter scales an object up from zero at its center while fading it in.
func GrowFromCenter(target mobject.Mobject) Animation {
	var from geometry.Vec2
	var fromOpacity float64
	return New(1, Linear, func(float64) {
		from = target.Base().Transform.Scale
		fromOpacity = target.Base().Opacity
		target.Base().IsVisible = true
	}, func(t float64) {
		target.Base().Transform.Scale = geometry.V2(from.X*t, from.Y*t)
		target.Base().Opacity = fromOpacity * t
	}, nil, func(float64) {
		target.Base().Transform.Scale = from
		target.Base().Opacity = fromOpacity
	})
}

// ShrinkToCenter scales an object down to zero at its center while fading it out.
func ShrinkToCenter(target mobject.Mobject) Animation {
	var from geometry.Vec2
	var fromOpacity float64
	return New(1, Linear, func(float64) {
		from = target.Base().Transform.Scale
		fromOpacity = target.Base().Opacity
	}, func(t float64) {
		target.Base().Transform.Scale = geometry.V2(from.X*(1-t), from.Y*(1-t))
		target.Base().Opacity = fromOpacity * (1 - t)
	}, func(float64) {
		target.Base().IsVisible = false
	}, func(float64) {
		target.Base().Transform.Scale = from
		target.Base().Opacity = fromOpacity
		target.Base().IsVisible = true
	})
}

// GrowFromPoint scales an object up from a fixed point, moving it back to its
// resting position as it grows.
func GrowFromPoint(target mobject.Mobject, point geometry.Vec2) Animation {
	var from, fromPos geometry.Vec2
	var fromOpacity float64
	return New(1, Linear, func(float64) {
		from = target.Base().Transform.Scale
		fromPos = target.Base().Transform.Position
		fromOpacity = target.Base().Opacity
		target.Base().IsVisible = true
	}, func(t float64) {
		target.Base().Transform.Scale = geometry.V2(from.X*t, from.Y*t)
		target.Base().Transform.Position = point.Lerp(fromPos, t)
		target.Base().Opacity = fromOpacity * t
	}, nil, func(float64) {
		target.Base().Transform.Scale = from
		target.Base().Transform.Position = fromPos
		target.Base().Opacity = fromOpacity
	})
}

// GrowFromEdge grows an object in from the direction given by edge
// ("left", "right", "up", or "down"), starting distance units away.
func GrowFromEdge(target mobject.Mobject, edge string, distance float64) Animation {
	if distance <= 0 {
		distance = 150
	}
	dir := edgeVector(edge)
	var from, fromPos geometry.Vec2
	var fromOpacity float64
	return New(1, Linear, func(float64) {
		from = target.Base().Transform.Scale
		fromPos = target.Base().Transform.Position
		fromOpacity = target.Base().Opacity
		target.Base().IsVisible = true
	}, func(t float64) {
		start := fromPos.Add(dir.Mul(distance))
		target.Base().Transform.Scale = geometry.V2(from.X*t, from.Y*t)
		target.Base().Transform.Position = start.Lerp(fromPos, t)
		target.Base().Opacity = fromOpacity * t
	}, nil, func(float64) {
		target.Base().Transform.Scale = from
		target.Base().Transform.Position = fromPos
		target.Base().Opacity = fromOpacity
	})
}

// SpinIn spins an object in from a full turn while scaling it up and fading it in.
func SpinIn(target mobject.Mobject) Animation {
	var from geometry.Vec2
	var fromRot, fromOpacity float64
	return New(1, Linear, func(float64) {
		from = target.Base().Transform.Scale
		fromRot = target.Base().Transform.Rotation
		fromOpacity = target.Base().Opacity
		target.Base().IsVisible = true
	}, func(t float64) {
		target.Base().Transform.Scale = geometry.V2(from.X*t, from.Y*t)
		target.Base().Transform.Rotation = fromRot - 2*math.Pi*(1-t)
		target.Base().Opacity = fromOpacity * t
	}, nil, func(float64) {
		target.Base().Transform.Scale = from
		target.Base().Transform.Rotation = fromRot
		target.Base().Opacity = fromOpacity
	})
}

// SpinOut spins an object out with a full turn while scaling it down and fading it out.
func SpinOut(target mobject.Mobject) Animation {
	var from geometry.Vec2
	var fromRot, fromOpacity float64
	return New(1, Linear, func(float64) {
		from = target.Base().Transform.Scale
		fromRot = target.Base().Transform.Rotation
		fromOpacity = target.Base().Opacity
	}, func(t float64) {
		target.Base().Transform.Scale = geometry.V2(from.X*(1-t), from.Y*(1-t))
		target.Base().Transform.Rotation = fromRot + 2*math.Pi*t
		target.Base().Opacity = fromOpacity * (1 - t)
	}, func(float64) {
		target.Base().IsVisible = false
	}, func(float64) {
		target.Base().Transform.Scale = from
		target.Base().Transform.Rotation = fromRot
		target.Base().Opacity = fromOpacity
		target.Base().IsVisible = true
	})
}

// Flash briefly brightens an object's fill and stroke toward a highlight color.
func Flash(target mobject.Mobject, highlight color.RGBA) Animation {
	var fromFill, fromStroke color.RGBA
	return New(1, Linear, func(float64) {
		fromFill = target.Base().Fill
		fromStroke = target.Base().Stroke
		target.Base().IsVisible = true
	}, func(t float64) {
		k := math.Sin(math.Pi * t)
		target.Base().Fill = lerpColor(fromFill, highlight, k)
		target.Base().Stroke = lerpColor(fromStroke, highlight, k)
	}, nil, func(float64) {
		target.Base().Fill = fromFill
		target.Base().Stroke = fromStroke
	})
}

// Indicate draws attention to an object by briefly scaling it up and tinting it.
func Indicate(target mobject.Mobject, scale float64, highlight color.RGBA) Animation {
	var from geometry.Vec2
	var fromFill, fromStroke color.RGBA
	return New(1, Linear, func(float64) {
		from = target.Base().Transform.Scale
		fromFill = target.Base().Fill
		fromStroke = target.Base().Stroke
	}, func(t float64) {
		k := math.Sin(math.Pi * t)
		s := 1 + (scale-1)*k
		target.Base().Transform.Scale = geometry.V2(from.X*s, from.Y*s)
		target.Base().Fill = lerpColor(fromFill, highlight, k)
		target.Base().Stroke = lerpColor(fromStroke, highlight, k)
	}, nil, func(float64) {
		target.Base().Transform.Scale = from
		target.Base().Fill = fromFill
		target.Base().Stroke = fromStroke
	})
}

// Blink dips an object's opacity to zero the given number of times.
func Blink(target mobject.Mobject, times float64) Animation {
	var fromOpacity float64
	return New(1, Linear, func(float64) {
		fromOpacity = target.Base().Opacity
		target.Base().IsVisible = true
	}, func(t float64) {
		k := math.Abs(math.Cos(math.Pi * times * t))
		target.Base().Opacity = fromOpacity * k
	}, nil, func(float64) {
		target.Base().Opacity = fromOpacity
	})
}

// Wiggle wobbles an object's rotation around its resting angle.
func Wiggle(target mobject.Mobject, amplitude, cycles float64) Animation {
	var fromRot float64
	return New(1, Linear, func(float64) {
		fromRot = target.Base().Transform.Rotation
	}, func(t float64) {
		target.Base().Transform.Rotation = fromRot + amplitude*math.Sin(2*math.Pi*cycles*t)
	}, nil, func(float64) {
		target.Base().Transform.Rotation = fromRot
	})
}

// Wave moves the direct children of a group (or a single object) up and down
// with a phase offset, producing a traveling wave.
func Wave(target mobject.Mobject, amplitude float64) Animation {
	children := waveChildren(target)
	positions := make([]geometry.Vec2, len(children))
	return New(1, Linear, func(float64) {
		for i, c := range children {
			positions[i] = c.Base().Transform.Position
		}
	}, func(t float64) {
		n := len(children)
		for i, c := range children {
			phase := 0.0
			if n > 1 {
				phase = float64(i) / float64(n-1) * 2 * math.Pi
			}
			offset := amplitude * math.Sin(2*math.Pi*t-phase)
			c.Base().Transform.Position = positions[i].Add(geometry.V2(0, offset))
		}
	}, nil, func(float64) {
		for i, c := range children {
			c.Base().Transform.Position = positions[i]
		}
	})
}

// ZoomIn scales an object down from an enlarged size into place while fading it in.
func ZoomIn(target mobject.Mobject, fromScale float64) Animation {
	var to geometry.Vec2
	var fromOpacity float64
	return New(1, Linear, func(float64) {
		to = target.Base().Transform.Scale
		fromOpacity = target.Base().Opacity
		target.Base().IsVisible = true
	}, func(t float64) {
		s := fromScale + (1-fromScale)*t
		target.Base().Transform.Scale = geometry.V2(to.X*s, to.Y*s)
		target.Base().Opacity = fromOpacity * t
	}, nil, func(float64) {
		target.Base().Transform.Scale = to
		target.Base().Opacity = fromOpacity
	})
}

// ZoomOut scales an object up to an enlarged size while fading it out.
func ZoomOut(target mobject.Mobject, toScale float64) Animation {
	var from geometry.Vec2
	var fromOpacity float64
	return New(1, Linear, func(float64) {
		from = target.Base().Transform.Scale
		fromOpacity = target.Base().Opacity
	}, func(t float64) {
		s := 1 + (toScale-1)*t
		target.Base().Transform.Scale = geometry.V2(from.X*s, from.Y*s)
		target.Base().Opacity = fromOpacity * (1 - t)
	}, func(float64) {
		target.Base().IsVisible = false
	}, func(float64) {
		target.Base().Transform.Scale = from
		target.Base().Opacity = fromOpacity
		target.Base().IsVisible = true
	})
}

// FocusOn briefly zooms an object in and back out to draw attention to it.
func FocusOn(target mobject.Mobject, scale float64) Animation {
	var from geometry.Vec2
	return New(1, Linear, func(float64) {
		from = target.Base().Transform.Scale
	}, func(t float64) {
		k := math.Sin(math.Pi * t)
		s := 1 + (scale-1)*k
		target.Base().Transform.Scale = geometry.V2(from.X*s, from.Y*s)
	}, nil, func(float64) {
		target.Base().Transform.Scale = from
	})
}

func edgeVector(edge string) geometry.Vec2 {
	switch edge {
	case "left":
		return geometry.V2(-1, 0)
	case "right":
		return geometry.V2(1, 0)
	case "up", "top":
		return geometry.V2(0, -1)
	case "down", "bottom":
		return geometry.V2(0, 1)
	default:
		return geometry.V2(0, 0)
	}
}

func waveChildren(target mobject.Mobject) []mobject.Mobject {
	if g, ok := target.(*mobject.Group); ok && len(g.Children) > 0 {
		return g.Children
	}
	return []mobject.Mobject{target}
}
