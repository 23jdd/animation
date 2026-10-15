package geometry

import "math"

// Matrix is an affine 2D matrix [A C E; B D F; 0 0 1].
type Matrix struct{ A, B, C, D, E, F float64 }

func Identity() Matrix        { return Matrix{A: 1, D: 1} }
func Translate(v Vec2) Matrix { m := Identity(); m.E, m.F = v.X, v.Y; return m }
func Scale(v Vec2) Matrix     { return Matrix{A: v.X, D: v.Y} }
func Rotate(angle float64) Matrix {
	c, s := math.Cos(angle), math.Sin(angle)
	return Matrix{A: c, B: s, C: -s, D: c}
}
func (m Matrix) Mul(n Matrix) Matrix {
	return Matrix{
		A: m.A*n.A + m.C*n.B, B: m.B*n.A + m.D*n.B,
		C: m.A*n.C + m.C*n.D, D: m.B*n.C + m.D*n.D,
		E: m.A*n.E + m.C*n.F + m.E, F: m.B*n.E + m.D*n.F + m.F,
	}
}
func (m Matrix) Apply(v Vec2) Vec2 { return V2(m.A*v.X+m.C*v.Y+m.E, m.B*v.X+m.D*v.Y+m.F) }
