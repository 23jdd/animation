package geometry

import "math"

// Vec3 is a renderer-independent three-dimensional vector.
type Vec3 struct{ X, Y, Z float64 }

func V3(x, y, z float64) Vec3     { return Vec3{X: x, Y: y, Z: z} }
func (v Vec3) Add(o Vec3) Vec3    { return V3(v.X+o.X, v.Y+o.Y, v.Z+o.Z) }
func (v Vec3) Sub(o Vec3) Vec3    { return V3(v.X-o.X, v.Y-o.Y, v.Z-o.Z) }
func (v Vec3) Mul(s float64) Vec3 { return V3(v.X*s, v.Y*s, v.Z*s) }
func (v Vec3) Dot(o Vec3) float64 { return v.X*o.X + v.Y*o.Y + v.Z*o.Z }
func (v Vec3) Cross(o Vec3) Vec3  { return V3(v.Y*o.Z-v.Z*o.Y, v.Z*o.X-v.X*o.Z, v.X*o.Y-v.Y*o.X) }
func (v Vec3) Length() float64    { return math.Sqrt(v.Dot(v)) }
func (v Vec3) Lerp(o Vec3, t float64) Vec3 {
	return V3(v.X+(o.X-v.X)*t, v.Y+(o.Y-v.Y)*t, v.Z+(o.Z-v.Z)*t)
}
func (v Vec3) Normalize() Vec3 {
	if length := v.Length(); length > 1e-12 {
		return v.Mul(1 / length)
	}
	return Vec3{}
}

// RotateEuler applies X, then Y, then Z rotation in radians.
func (v Vec3) RotateEuler(rotation Vec3) Vec3 {
	cx, sx := math.Cos(rotation.X), math.Sin(rotation.X)
	cy, sy := math.Cos(rotation.Y), math.Sin(rotation.Y)
	cz, sz := math.Cos(rotation.Z), math.Sin(rotation.Z)

	x1, y1, z1 := v.X, v.Y*cx-v.Z*sx, v.Y*sx+v.Z*cx
	x2, y2, z2 := x1*cy+z1*sy, y1, -x1*sy+z1*cy
	return V3(x2*cz-y2*sz, x2*sz+y2*cz, z2)
}

type Transform3D struct {
	Position Vec3
	Rotation Vec3
	Scale    Vec3
}

func NewTransform3D() Transform3D { return Transform3D{Scale: V3(1, 1, 1)} }

func (t Transform3D) Apply(v Vec3) Vec3 {
	v = V3(v.X*t.Scale.X, v.Y*t.Scale.Y, v.Z*t.Scale.Z)
	return v.RotateEuler(t.Rotation).Add(t.Position)
}
