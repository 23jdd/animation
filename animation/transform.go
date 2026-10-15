package animation

import (
	"image/color"
	"math"

	"github.com/23dd/animation/geometry"
	"github.com/23dd/animation/mobject"
)

func MoveTo(target mobject.Mobject, to geometry.Vec2) Animation {
	var from geometry.Vec2
	return New(1, Linear, func(float64) { from = target.Base().Transform.Position }, func(t float64) { target.Base().Transform.Position = from.Lerp(to, t) }, nil, func(float64) { target.Base().Transform.Position = from })
}
func MoveBy(target mobject.Mobject, delta geometry.Vec2) Animation {
	var from geometry.Vec2
	return New(1, Linear, func(float64) { from = target.Base().Transform.Position }, func(t float64) { target.Base().Transform.Position = from.Add(delta.Mul(t)) }, nil, func(float64) { target.Base().Transform.Position = from })
}
func ScaleTo(target mobject.Mobject, to geometry.Vec2) Animation {
	var from geometry.Vec2
	return New(1, Linear, func(float64) { from = target.Base().Transform.Scale }, func(t float64) { target.Base().Transform.Scale = from.Lerp(to, t) }, nil, func(float64) { target.Base().Transform.Scale = from })
}
func ScaleBy(target mobject.Mobject, factor geometry.Vec2) Animation {
	var from geometry.Vec2
	return New(1, Linear, func(float64) { from = target.Base().Transform.Scale }, func(t float64) {
		target.Base().Transform.Scale = geometry.V2(from.X*(1+(factor.X-1)*t), from.Y*(1+(factor.Y-1)*t))
	}, nil, func(float64) { target.Base().Transform.Scale = from })
}
func RotateTo(target mobject.Mobject, angle float64) Animation {
	var from float64
	return New(1, Linear, func(float64) { from = target.Base().Transform.Rotation }, func(t float64) { target.Base().Transform.Rotation = from + (angle-from)*t }, nil, func(float64) { target.Base().Transform.Rotation = from })
}
func RotateBy(target mobject.Mobject, angle float64) Animation {
	var from float64
	return New(1, Linear, func(float64) { from = target.Base().Transform.Rotation }, func(t float64) { target.Base().Transform.Rotation = from + angle*t }, nil, func(float64) { target.Base().Transform.Rotation = from })
}
func OpacityTo(target mobject.Mobject, opacity float64) Animation {
	var from float64
	return New(1, Linear, func(float64) { from = target.Base().Opacity }, func(t float64) { target.Base().Opacity = from + (opacity-from)*t }, nil, func(float64) { target.Base().Opacity = from })
}
func FadeIn(target mobject.Mobject) Animation {
	var from float64
	return New(1, Linear, func(float64) { from = target.Base().Opacity; target.Base().IsVisible = true }, func(t float64) { target.Base().Opacity = t }, nil, func(float64) { target.Base().Opacity = from })
}
func FadeOut(target mobject.Mobject) Animation {
	var from float64
	return New(1, Linear, func(float64) { from = target.Base().Opacity }, func(t float64) { target.Base().Opacity = from * (1 - t) }, func(float64) { target.Base().IsVisible = false }, func(float64) { target.Base().Opacity = from; target.Base().IsVisible = true })
}
func Create(target mobject.Mobject) Animation {
	var from float64
	return New(1, Linear, func(float64) { from = target.Base().DrawProgress; target.Base().IsVisible = true }, func(t float64) { target.Base().DrawProgress = t }, nil, func(float64) { target.Base().DrawProgress = from })
}
func Uncreate(target mobject.Mobject) Animation {
	var from float64
	return New(1, Linear, func(float64) { from = target.Base().DrawProgress }, func(t float64) { target.Base().DrawProgress = from * (1 - t) }, func(float64) { target.Base().IsVisible = false }, func(float64) { target.Base().DrawProgress = from; target.Base().IsVisible = true })
}
func ColorTo(target mobject.Mobject, to color.RGBA) Animation {
	var from color.RGBA
	lerp := func(a, b uint8, t float64) uint8 { return uint8(math.Round(float64(a) + (float64(b)-float64(a))*t)) }
	return New(1, Linear, func(float64) { from = target.Base().Fill }, func(t float64) {
		target.Base().Fill = color.RGBA{lerp(from.R, to.R, t), lerp(from.G, to.G, t), lerp(from.B, to.B, t), lerp(from.A, to.A, t)}
	}, nil, func(float64) { target.Base().Fill = from })
}
func MoveAlongPath(target mobject.Mobject, path geometry.Path) Animation {
	points := flattenPath(path, 24)
	var from geometry.Vec2
	return New(1, Linear, func(float64) { from = target.Base().Transform.Position }, func(t float64) {
		if len(points) > 0 {
			target.Base().Transform.Position = from.Add(pointAt(points, t))
		}
	}, nil, func(float64) { target.Base().Transform.Position = from })
}

// Transform interpolates renderer-independent state shared by all mobjects.
func Transform(from, to mobject.Mobject) Animation {
	var pos, scale geometry.Vec2
	var rotation, opacity float64
	var fill color.RGBA
	lerp := func(a, b uint8, t float64) uint8 { return uint8(math.Round(float64(a) + (float64(b)-float64(a))*t)) }
	return New(1, Linear, func(float64) {
		b := from.Base()
		pos, scale, rotation, opacity, fill = b.Transform.Position, b.Transform.Scale, b.Transform.Rotation, b.Opacity, b.Fill
	}, func(t float64) {
		a, b := from.Base(), to.Base()
		a.Transform.Position = pos.Lerp(b.Transform.Position, t)
		a.Transform.Scale = scale.Lerp(b.Transform.Scale, t)
		a.Transform.Rotation = rotation + (b.Transform.Rotation-rotation)*t
		a.Opacity = opacity + (b.Opacity-opacity)*t
		a.Fill = color.RGBA{lerp(fill.R, b.Fill.R, t), lerp(fill.G, b.Fill.G, t), lerp(fill.B, b.Fill.B, t), lerp(fill.A, b.Fill.A, t)}
	}, nil, func(float64) {
		b := from.Base()
		b.Transform.Position, b.Transform.Scale, b.Transform.Rotation, b.Opacity, b.Fill = pos, scale, rotation, opacity, fill
	})
}

func ReplacementTransform(from, to mobject.Mobject) Animation {
	inner := Transform(from, to)
	return New(inner.Duration(), Linear, func(float64) { inner.Start() }, inner.Update, func(float64) {
		inner.Finish()
		from.Base().IsVisible, to.Base().IsVisible = false, true
	}, func(float64) {
		inner.Reset()
		from.Base().IsVisible, to.Base().IsVisible = true, false
	})
}
func flattenPath(path geometry.Path, steps int) []geometry.Vec2 {
	var out []geometry.Vec2
	var cur, start geometry.Vec2
	for _, c := range path.Commands {
		switch c.Type {
		case geometry.MoveCommand:
			cur = c.Points[0]
			start = cur
			out = append(out, cur)
		case geometry.LineCommand:
			cur = c.Points[0]
			out = append(out, cur)
		case geometry.QuadraticCommand:
			a := cur
			for i := 1; i <= steps; i++ {
				out = append(out, geometry.QuadraticBezier(a, c.Points[0], c.Points[1], float64(i)/float64(steps)))
			}
			cur = c.Points[1]
		case geometry.CubicCommand:
			a := cur
			for i := 1; i <= steps; i++ {
				out = append(out, geometry.CubicBezier(a, c.Points[0], c.Points[1], c.Points[2], float64(i)/float64(steps)))
			}
			cur = c.Points[2]
		case geometry.CloseCommand:
			cur = start
			out = append(out, cur)
		}
	}
	return out
}
func pointAt(p []geometry.Vec2, t float64) geometry.Vec2 {
	if len(p) == 1 {
		return p[0]
	}
	total := 0.0
	for i := 1; i < len(p); i++ {
		total += p[i].Sub(p[i-1]).Length()
	}
	d := total * clamp(t)
	for i := 1; i < len(p); i++ {
		s := p[i].Sub(p[i-1]).Length()
		if d <= s {
			if s == 0 {
				return p[i]
			}
			return p[i-1].Lerp(p[i], d/s)
		}
		d -= s
	}
	return p[len(p)-1]
}
