package gg

import (
	"image"
	"image/color"
	"math"
	"os"

	"github.com/23dd/animation/fontasset"
	"github.com/23dd/animation/geometry"
	"github.com/23dd/animation/mobject"
	"github.com/23dd/animation/scene"
	base "github.com/fogleman/gg"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

type Renderer struct {
	width, height             int
	outputWidth, outputHeight int
	outputScale               float64
	dc                        *base.Context
	primaryFont               *opentype.Font
	fallbackFont              *opentype.Font
	fonts                     map[fontSourceKey]*opentype.Font
	faces                     map[fontFaceKey]font.Face
}

type fontSourceKey struct {
	Path    string
	Version int64
}
type fontFaceKey struct {
	Source fontSourceKey
	Size   int
}

func New(width, height int) *Renderer {
	primary, err := opentype.Parse(fontasset.JetBrainsMonoRegular())
	if err != nil {
		panic("parse embedded JetBrains Mono: " + err.Error())
	}
	fallback, err := opentype.Parse(fontasset.SourceHanSansSCRegular())
	if err != nil {
		panic("parse embedded Source Han Sans SC fallback: " + err.Error())
	}
	return &Renderer{width: width, height: height, outputWidth: width, outputHeight: height, outputScale: 1, dc: base.NewContext(width, height), primaryFont: primary, fallbackFont: fallback, fonts: make(map[fontSourceKey]*opentype.Font), faces: make(map[fontFaceKey]font.Face)}
}
func (r *Renderer) Size() (int, int) { return r.width, r.height }

// OutputSize reports raster dimensions independently of logical scene dimensions.
func (r *Renderer) OutputSize() (int, int) { return r.outputWidth, r.outputHeight }

// SetOutputSize keeps scene coordinates logical while changing raster output.
func (r *Renderer) SetOutputSize(width, height int) {
	if width < 1 || height < 1 {
		return
	}
	if width == r.outputWidth && height == r.outputHeight {
		return
	}
	r.outputWidth, r.outputHeight = width, height
	r.outputScale = math.Min(float64(width)/float64(r.width), float64(height)/float64(r.height))
	r.dc = base.NewContext(width, height)
}

func (r *Renderer) Render(s *scene.Scene) image.Image {
	bg := s.Camera.Background
	r.dc.SetRGBA255(int(bg.R), int(bg.G), int(bg.B), int(bg.A))
	r.dc.Clear()
	r.dc.Push()
	r.dc.Translate(float64(r.outputWidth)/2-s.Camera.Position.X*s.Camera.Zoom*r.outputScale, float64(r.outputHeight)/2-s.Camera.Position.Y*s.Camera.Zoom*r.outputScale)
	r.dc.Scale(s.Camera.Zoom*r.outputScale, s.Camera.Zoom*r.outputScale)
	for _, o := range s.Snapshot() {
		r.draw(o, 1)
	}
	r.dc.Pop()
	return r.dc.Image()
}
func withAlpha(c color.RGBA, a float64) color.RGBA {
	c.A = uint8(math.Round(float64(c.A) * math.Max(0, math.Min(1, a))))
	return c
}
func (r *Renderer) setColor(c color.RGBA) { r.dc.SetRGBA255(int(c.R), int(c.G), int(c.B), int(c.A)) }
func (r *Renderer) draw(o mobject.Mobject, parentAlpha float64) {
	b := o.Base()
	if !b.IsVisible || b.Opacity <= 0 {
		return
	}
	alpha := parentAlpha * b.Opacity
	r.dc.Push()
	r.dc.Translate(b.Transform.Position.X, b.Transform.Position.Y)
	r.dc.Rotate(b.Transform.Rotation)
	r.dc.Scale(b.Transform.Scale.X, b.Transform.Scale.Y)
	if g, ok := o.(*mobject.Group); ok {
		for _, child := range g.Children {
			r.draw(child, alpha)
		}
		r.dc.Pop()
		return
	}
	switch v := o.(type) {
	case *mobject.Circle:
		if b.DrawProgress < 1 {
			r.dc.DrawArc(0, 0, v.Radius, -math.Pi/2, -math.Pi/2+2*math.Pi*math.Max(0, b.DrawProgress))
			r.stroke(b, alpha)
		} else {
			r.dc.DrawCircle(0, 0, v.Radius)
			r.paint(b, alpha)
		}
	case *mobject.Rectangle:
		x, y := -v.Width/2, -v.Height/2
		if v.CornerRadius > 0 {
			r.dc.DrawRoundedRectangle(x, y, v.Width, v.Height, v.CornerRadius)
		} else {
			r.dc.DrawRectangle(x, y, v.Width, v.Height)
		}
		r.paint(b, alpha)
	case *mobject.Line:
		end := v.Start.Lerp(v.End, b.DrawProgress)
		r.dc.MoveTo(v.Start.X, v.Start.Y)
		r.dc.LineTo(end.X, end.Y)
		r.stroke(b, alpha)
	case *mobject.Polygon:
		if len(v.Points) > 0 {
			n := int(math.Ceil(float64(len(v.Points)) * b.DrawProgress))
			if n < 1 {
				n = 1
			}
			if n > len(v.Points) {
				n = len(v.Points)
			}
			r.dc.MoveTo(v.Points[0].X, v.Points[0].Y)
			for _, p := range v.Points[1:n] {
				r.dc.LineTo(p.X, p.Y)
			}
			if b.DrawProgress >= 1 {
				r.dc.ClosePath()
			}
			r.paint(b, alpha)
		}
	case *mobject.Path:
		r.path(v.Geometry, b.DrawProgress)
		r.paint(b, alpha)
	case *mobject.Text:
		r.setColor(withAlpha(b.Fill, alpha))
		size := int(math.Round(v.Size * r.outputScale))
		if size < 1 {
			size = 1
		}
		face := r.fontFace(v.Font, v.FontVersion, size)
		if face != nil {
			r.dc.SetFontFace(face)
		}
		r.dc.Push()
		r.dc.Scale(1/r.outputScale, 1/r.outputScale)
		r.dc.DrawStringAnchored(v.Content, 0, 0, .5, .5)
		r.dc.Pop()
	}
	r.dc.Pop()
}

func (r *Renderer) fontFace(path string, version int64, size int) font.Face {
	sourceKey := fontSourceKey{Path: path, Version: version}
	faceKey := fontFaceKey{Source: sourceKey, Size: size}
	if face := r.faces[faceKey]; face != nil {
		return face
	}
	if path == "" {
		primary := newOpenTypeFace(r.primaryFont, size)
		fallback := newOpenTypeFace(r.fallbackFont, size)
		if primary == nil || fallback == nil {
			return nil
		}
		face := &fallbackFace{primary: primary, fallback: fallback}
		r.faces[faceKey] = face
		return face
	}
	var parsed *opentype.Font
	if path != "" {
		if customFont := r.fonts[sourceKey]; customFont != nil {
			parsed = customFont
		} else {
			if data, err := os.ReadFile(path); err == nil {
				customFont, parseErr := opentype.Parse(data)
				if parseErr == nil {
					parsed = customFont
					r.fonts[sourceKey] = customFont
				}
			}
		}
	}
	if parsed == nil {
		return nil
	}
	face := newOpenTypeFace(parsed, size)
	if face == nil {
		return nil
	}
	r.faces[faceKey] = face
	return face
}

func newOpenTypeFace(parsed *opentype.Font, size int) font.Face {
	face, err := opentype.NewFace(parsed, &opentype.FaceOptions{Size: float64(size), DPI: 72, Hinting: font.HintingFull})
	if err != nil {
		return nil
	}
	return face
}
func (r *Renderer) paint(b *mobject.BaseMobject, alpha float64) {
	if b.Fill.A > 0 {
		r.setColor(withAlpha(b.Fill, alpha))
		r.dc.FillPreserve()
	}
	if b.Stroke.A > 0 && b.StrokeWidth > 0 {
		r.stroke(b, alpha)
	} else {
		r.dc.NewSubPath()
	}
}
func (r *Renderer) stroke(b *mobject.BaseMobject, alpha float64) {
	r.setColor(withAlpha(b.Stroke, alpha))
	r.dc.SetLineWidth(b.StrokeWidth * r.outputScale)
	r.dc.Stroke()
}
func (r *Renderer) path(p geometry.Path, progress float64) {
	limit := int(math.Ceil(float64(len(p.Commands)) * math.Max(0, math.Min(1, progress))))
	for i, c := range p.Commands {
		if i >= limit {
			break
		}
		switch c.Type {
		case geometry.MoveCommand:
			r.dc.MoveTo(c.Points[0].X, c.Points[0].Y)
		case geometry.LineCommand:
			r.dc.LineTo(c.Points[0].X, c.Points[0].Y)
		case geometry.QuadraticCommand:
			r.dc.QuadraticTo(c.Points[0].X, c.Points[0].Y, c.Points[1].X, c.Points[1].Y)
		case geometry.CubicCommand:
			r.dc.CubicTo(c.Points[0].X, c.Points[0].Y, c.Points[1].X, c.Points[1].Y, c.Points[2].X, c.Points[2].Y)
		case geometry.CloseCommand:
			r.dc.ClosePath()
		}
	}
}
