package geometry

type Transform struct {
	Position Vec2
	Rotation float64
	Scale    Vec2
}

func NewTransform() Transform { return Transform{Scale: V2(1, 1)} }
func (t Transform) Matrix() Matrix {
	return Translate(t.Position).Mul(Rotate(t.Rotation)).Mul(Scale(t.Scale))
}
