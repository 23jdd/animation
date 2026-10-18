package gg

import (
	"errors"
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
	initErr                   error
	camera                    *scene.Camera
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
	primaryErr := err
	fallback, err := opentype.Parse(fontasset.SourceHanSansSCRegular())
	fallbackErr := err
	return &Renderer{width: width, height: height, outputWidth: width, outputHeight: height, outputScale: 1, dc: base.NewContext(width, height), primaryFont: primary, fallbackFont: fallback, fonts: make(map[fontSourceKey]*opentype.Font), faces: make(map[fontFaceKey]font.Face), initErr: errors.Join(primaryErr, fallbackErr)}
}
func (r *Renderer) Size() (int, int) { return r.width, r.height }

// Err reports initialization failures such as invalid embedded font data.
func (r *Renderer) Err() error { return r.initErr }

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
	r.camera = s.Camera
	bg := s.Camera.Background
	r.dc.SetRGBA255(int(bg.R), int(bg.G), int(bg.B), int(bg.A))
	r.dc.Clear()
	r.dc.Push()
	r.dc.Translate(float64(r.outputWidth)/2-s.Camera.Position.X*s.Camera.Zoom*r.outputScale, float64(r.outputHeight)/2-s.Camera.Position.Y*s.Camera.Zoom*r.outputScale)
	r.dc.Scale(s.Camera.Zoom*r.outputScale, s.Camera.Zoom*r.outputScale)
	objects := s.Snapshot()
	for i := 0; i < len(objects); {
		if _, ok := objects[i].(*mobject.Mesh3D); ok {
			var meshes []*mobject.Mesh3D
			for i < len(objects) {
				mesh, meshOK := objects[i].(*mobject.Mesh3D)
				if !meshOK {
					break
				}
				meshes = append(meshes, mesh)
				i++
			}
			r.drawMeshes3D(meshes, 1, true)
			continue
		}
		r.draw(objects[i], 1)
		i++
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
	case *mobject.Image:
		r.drawImage(v, alpha)
	case *mobject.Mesh3D:
		r.drawMeshes3D([]*mobject.Mesh3D{v}, alpha, false)
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
			r.paint(b, alpha)
		} else if b.DrawProgress < 1 {
			r.rectangleStrokeProgress(x, y, v.Width, v.Height, math.Max(0, b.DrawProgress))
			r.strokeRect(b, alpha)
		} else {
			r.dc.DrawRectangle(x, y, v.Width, v.Height)
			r.paint(b, alpha)
		}
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
		size := int(math.Round(v.Size * r.outputScale))
		if size < 1 {
			size = 1
		}
		face := r.fontFace(v.Font, v.FontVersion, size)
		if face != nil {
			r.dc.SetFontFace(face)
		}
		runes := []rune(v.Content)
		total := len(runes)
		reveal := math.Max(0, math.Min(1, v.Reveal))
		if total == 0 || reveal <= 0 {
			break
		}
		full := int(math.Floor(reveal * float64(total)))
		frac := reveal*float64(total) - float64(full)
		if full >= total {
			full = total
			frac = 0
		}
		fullWidth, _ := r.dc.MeasureString(v.Content)
		prefixWidth := 0.0
		if full > 0 {
			prefixWidth, _ = r.dc.MeasureString(string(runes[:full]))
		}
		currentWidth := 0.0
		if full < total {
			currentWidth, _ = r.dc.MeasureString(string(runes[full]))
		}
		r.dc.Push()
		r.dc.Scale(1/r.outputScale, 1/r.outputScale)
		// Fully revealed characters are drawn at full opacity.
		if full > 0 {
			r.setColor(withAlpha(b.Fill, alpha))
			r.dc.DrawStringAnchored(string(runes[:full]), (prefixWidth-fullWidth)/2, 0, .5, .5)
		}
		// The current character fades in as the reveal value advances past it.
		if full < total && frac > 0 {
			cx := prefixWidth + currentWidth/2 - fullWidth/2
			r.setColor(withAlpha(b.Fill, alpha*frac))
			r.dc.DrawStringAnchored(string(runes[full]), cx, 0, .5, .5)
		}
		// Blinking typewriter caret follows the last revealed character.
		if v.Cursor.A > 0 && v.CursorVisible && full < total {
			cursorX := prefixWidth + currentWidth - fullWidth/2
			halfH := v.Size * r.outputScale / 2
			r.setColor(withAlpha(v.Cursor, alpha))
			r.dc.SetLineWidth(math.Max(1, v.Size*0.06) * r.outputScale)
			r.dc.DrawLine(cursorX, -halfH, cursorX, halfH)
			r.dc.Stroke()
		}
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
	if parsed == nil {
		return nil
	}
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
		// FillPreserve leaves geometry behind; the next object must not reuse it.
		r.dc.ClearPath()
	}
}
func (r *Renderer) stroke(b *mobject.BaseMobject, alpha float64) {
	r.setColor(withAlpha(b.Stroke, alpha))
	r.dc.SetLineWidth(b.StrokeWidth * r.outputScale)
	r.dc.Stroke()
}

// strokeRect strokes the current path, falling back to the fill color when the
// stroke is invisible so that fill-only rectangles still reveal progressively.
func (r *Renderer) strokeRect(b *mobject.BaseMobject, alpha float64) {
	col, width := b.Stroke, b.StrokeWidth
	if col.A == 0 || width <= 0 {
		col = b.Fill
		if width <= 0 {
			width = 2
		}
	}
	if col.A == 0 {
		r.dc.ClearPath()
		return
	}
	r.setColor(withAlpha(col, alpha))
	r.dc.SetLineWidth(width * r.outputScale)
	r.dc.Stroke()
}

// rectangleStrokeProgress traces the perimeter of a rectangle clockwise from
// its top-left corner, up to the given fraction of the full perimeter.
func (r *Renderer) rectangleStrokeProgress(x, y, w, h, progress float64) {
	if progress <= 0 {
		return
	}
	if progress >= 1 {
		r.dc.DrawRectangle(x, y, w, h)
		return
	}
	remaining := 2 * (w + h) * progress
	corners := [4][2]float64{
		{x, y},
		{x + w, y},
		{x + w, y + h},
		{x, y + h},
	}
	r.dc.MoveTo(corners[0][0], corners[0][1])
	for i := 0; i < 4 && remaining > 0; i++ {
		a := corners[i]
		b := corners[(i+1)%4]
		seg := math.Hypot(b[0]-a[0], b[1]-a[1])
		if seg == 0 {
			continue
		}
		if remaining >= seg {
			r.dc.LineTo(b[0], b[1])
			remaining -= seg
		} else {
			t := remaining / seg
			r.dc.LineTo(a[0]+(b[0]-a[0])*t, a[1]+(b[1]-a[1])*t)
			remaining = 0
		}
	}
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
