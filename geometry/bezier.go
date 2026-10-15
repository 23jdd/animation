package geometry

func QuadraticBezier(a, b, c Vec2, t float64) Vec2 {
	u := 1 - t
	return a.Mul(u * u).Add(b.Mul(2 * u * t)).Add(c.Mul(t * t))
}

func CubicBezier(a, b, c, d Vec2, t float64) Vec2 {
	u := 1 - t
	return a.Mul(u * u * u).Add(b.Mul(3 * u * u * t)).Add(c.Mul(3 * u * t * t)).Add(d.Mul(t * t * t))
}
