package geometry

import "math"

type Vec2 struct{ X, Y float64 }

func V2(x, y float64) Vec2                 { return Vec2{X: x, Y: y} }
func (v Vec2) Add(o Vec2) Vec2             { return V2(v.X+o.X, v.Y+o.Y) }
func (v Vec2) Sub(o Vec2) Vec2             { return V2(v.X-o.X, v.Y-o.Y) }
func (v Vec2) Mul(s float64) Vec2          { return V2(v.X*s, v.Y*s) }
func (v Vec2) Length() float64             { return math.Hypot(v.X, v.Y) }
func (v Vec2) Lerp(o Vec2, t float64) Vec2 { return V2(v.X+(o.X-v.X)*t, v.Y+(o.Y-v.Y)*t) }
