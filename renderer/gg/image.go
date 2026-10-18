package gg

import (
	"image"
	"image/color"
	"math"

	"github.com/23dd/animation/mobject"
	"golang.org/x/image/draw"
	"golang.org/x/image/math/f64"
)

func (r *Renderer) drawImage(m *mobject.Image, alpha float64) {
	if m.Pixels == nil || !positiveFinite(m.Width) || !positiveFinite(m.Height) || math.IsNaN(alpha) {
		return
	}
	bounds := m.Pixels.Bounds()
	if bounds.Empty() {
		return
	}
	progress := math.Max(0, math.Min(1, m.DrawProgress))
	if progress <= 0 || math.IsNaN(progress) {
		return
	}
	// Read the complete object/group/camera/output transform from gg. A
	// source-space opacity mask avoids copying the bitmap for each fade frame.
	x0, y0 := r.dc.TransformPoint(0, 0)
	x1, y1 := r.dc.TransformPoint(1, 0)
	x2, y2 := r.dc.TransformPoint(0, 1)
	a, b, c, d := x1-x0, x2-x0, y1-y0, y2-y0
	det := a*d - b*c
	if det == 0 || math.IsNaN(det) || math.IsInf(det, 0) {
		return
	}
	sx, sy := m.Width/float64(bounds.Dx()), m.Height/float64(bounds.Dy())
	switch m.Fit {
	case mobject.ImageCover:
		sx = math.Max(sx, sy)
		sy = sx
	case mobject.ImageStretch:
	default:
		sx = math.Min(sx, sy)
		sy = sx
	}
	ox := -float64(bounds.Dx())*sx/2 - float64(bounds.Min.X)*sx
	oy := -float64(bounds.Dy())*sy/2 - float64(bounds.Min.Y)*sy
	matrix := f64.Aff3{a * sx, b * sy, x0 + a*ox + b*oy, c * sx, d * sy, y0 + c*ox + d*oy}
	for _, v := range matrix {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return
		}
	}
	opts := &draw.Options{}
	if alpha < 1 {
		opts.SrcMask = image.NewUniform(color.Alpha{A: uint8(math.Round(255 * math.Max(0, alpha)))})
	}
	if m.Fit == mobject.ImageCover || progress < 1 {
		opts.DstMask = &imageBoxMask{
			bounds: r.dc.Image().Bounds(),
			inverse: f64.Aff3{d / det, -b / det, (b*y0 - d*x0) / det,
				-c / det, a / det, (c*x0 - a*y0) / det},
			left: -m.Width / 2, right: -m.Width/2 + m.Width*progress,
			top: -m.Height / 2, bottom: m.Height / 2,
		}
	}
	draw.ApproxBiLinear.Transform(r.dc.Image().(draw.Image), matrix, m.Pixels, bounds, draw.Over, opts)
}

func positiveFinite(x float64) bool { return x > 0 && !math.IsNaN(x) && !math.IsInf(x, 0) }

// A procedural destination mask clips cover/reveal to the transformed logical
// box, including under rotation and mirroring. No full-frame mask is allocated.
type imageBoxMask struct {
	bounds                   image.Rectangle
	inverse                  f64.Aff3
	left, right, top, bottom float64
}

func (m *imageBoxMask) ColorModel() color.Model { return color.AlphaModel }
func (m *imageBoxMask) Bounds() image.Rectangle { return m.bounds }
func (m *imageBoxMask) At(x, y int) color.Color {
	if !image.Pt(x, y).In(m.bounds) {
		return color.Alpha{}
	}
	a := m.inverse
	px, py := float64(x)+.5, float64(y)+.5
	u, v := a[0]*px+a[1]*py+a[2], a[3]*px+a[4]*py+a[5]
	ex, ey := (math.Abs(a[0])+math.Abs(a[1]))/2, (math.Abs(a[3])+math.Abs(a[4]))/2
	if u-ex >= m.left && u+ex <= m.right && v-ey >= m.top && v+ey <= m.bottom {
		return color.Alpha{A: 255}
	}
	if u+ex <= m.left || u-ex >= m.right || v+ey <= m.top || v-ey >= m.bottom {
		return color.Alpha{}
	}
	// Four subpixel samples smooth only the transformed clipping edge.
	inside := 0
	for _, dx := range [...]float64{-.25, .25} {
		for _, dy := range [...]float64{-.25, .25} {
			x, y := u+a[0]*dx+a[1]*dy, v+a[3]*dx+a[4]*dy
			if x >= m.left && x < m.right && y >= m.top && y < m.bottom {
				inside++
			}
		}
	}
	return color.Alpha{A: uint8((inside*255 + 2) / 4)}
}
