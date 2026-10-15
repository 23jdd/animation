package mobject

import (
	"image/color"
	"math"

	"github.com/23dd/animation/geometry"
)

type Circle struct {
	BaseMobject
	Radius float64
}

func NewCircle(radius float64) *Circle { return &Circle{BaseMobject: NewBase(), Radius: radius} }
func (m *Circle) Update(t, dt float64) { Update(m, t, dt) }

type Rectangle struct {
	BaseMobject
	Width, Height, CornerRadius float64
}

func NewRectangle(width, height float64) *Rectangle {
	return &Rectangle{BaseMobject: NewBase(), Width: width, Height: height}
}
func (m *Rectangle) Update(t, dt float64) { Update(m, t, dt) }

type Line struct {
	BaseMobject
	Start, End geometry.Vec2
}

func NewLine(start, end geometry.Vec2) *Line {
	b := NewBase()
	b.Fill.A = 0
	return &Line{BaseMobject: b, Start: start, End: end}
}
func (m *Line) Update(t, dt float64) { Update(m, t, dt) }

type Polygon struct {
	BaseMobject
	Points []geometry.Vec2
}

func NewPolygon(points ...geometry.Vec2) *Polygon {
	return &Polygon{BaseMobject: NewBase(), Points: append([]geometry.Vec2(nil), points...)}
}
func (m *Polygon) Update(t, dt float64) { Update(m, t, dt) }

type Path struct {
	BaseMobject
	Geometry geometry.Path
}

func NewPath(path geometry.Path) *Path { return &Path{BaseMobject: NewBase(), Geometry: path} }
func (m *Path) Update(t, dt float64)   { Update(m, t, dt) }

type Text struct {
	BaseMobject
	Content     string
	Size        float64
	Font        string
	FontVersion int64
	Reveal      float64
	// Cursor, when non-zero, renders a blinking typewriter caret after the
	// last revealed character. CursorVisible is the per-frame blink state.
	Cursor        color.RGBA
	CursorVisible bool
}

func NewText(content string, size float64) *Text {
	return &Text{BaseMobject: NewBase(), Content: content, Size: size, Reveal: 1}
}
func (m *Text) Update(t, dt float64) {
	if m.Cursor.A > 0 {
		m.CursorVisible = math.Mod(t, 1.0) < 0.5
	} else {
		m.CursorVisible = false
	}
	Update(m, t, dt)
}
