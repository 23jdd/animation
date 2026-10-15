package geometry

type Rect struct{ Min, Max Vec2 }

func R(x, y, width, height float64) Rect { return Rect{V2(x, y), V2(x+width, y+height)} }
func (r Rect) Width() float64            { return r.Max.X - r.Min.X }
func (r Rect) Height() float64           { return r.Max.Y - r.Min.Y }
func (r Rect) Center() Vec2              { return r.Min.Lerp(r.Max, .5) }
