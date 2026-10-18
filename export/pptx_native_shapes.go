package export

import (
	"bytes"
	"errors"
	"fmt"
	"image/color"
	"math"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/23dd/animation/fontasset"
	"github.com/23dd/animation/geometry"
	"github.com/23dd/animation/mobject"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
)

type nativeSpec struct {
	left, top, width, height float64
	geometry, text, family   string
	fontSize                 float64
	fill, stroke             color.RGBA
	strokeWidth              float64
}

type nativeFont struct {
	font   *opentype.Font
	family string
}

type nativeTextKey struct {
	path, text string
	size       float64
}
type nativeTextLayout struct {
	width, height float64
	family        string
}
type nativeBuilder struct {
	fonts map[string]nativeFont
	text  map[nativeTextKey]nativeTextLayout
}

func newNativeBuilder() *nativeBuilder {
	return &nativeBuilder{fonts: make(map[string]nativeFont), text: make(map[nativeTextKey]nativeTextLayout)}
}

func (b *nativeBuilder) spec(obj mobject.Mobject) (nativeSpec, error) {
	base := obj.Base()
	s := nativeSpec{fill: base.Fill, stroke: base.Stroke, strokeWidth: base.StrokeWidth}
	switch o := obj.(type) {
	case *mobject.Circle:
		s.left, s.top, s.width, s.height = -o.Radius, -o.Radius, 2*o.Radius, 2*o.Radius
		s.geometry = `<a:prstGeom prst="ellipse"><a:avLst/></a:prstGeom>`
	case *mobject.Rectangle:
		s.left, s.top, s.width, s.height = -o.Width/2, -o.Height/2, o.Width, o.Height
		if o.CornerRadius > 0 {
			adjust := int(math.Round(math.Min(.5, o.CornerRadius/math.Min(o.Width, o.Height)) * 100000))
			s.geometry = fmt.Sprintf(`<a:prstGeom prst="roundRect"><a:avLst><a:gd name="adj" fmla="val %d"/></a:avLst></a:prstGeom>`, adjust)
		} else {
			s.geometry = `<a:prstGeom prst="rect"><a:avLst/></a:prstGeom>`
		}
	case *mobject.Text:
		if o.Reveal != 1 || o.Cursor.A != 0 {
			return s, errors.New("write/unwrite/cursor effects are not supported in native mode")
		}
		if !utf8.ValidString(o.Content) {
			return s, errors.New("native text must be valid UTF-8")
		}
		for _, r := range o.Content {
			if (r < 32 && r != '\n' && r != '\r' && r != '\t') || r == 0xfffe || r == 0xffff {
				return s, errors.New("native text contains an XML control character")
			}
		}
		if !nativeFinite(o.Size) || o.Size <= 0 || o.Size > 10000 {
			return s, errors.New("native text size must be positive, finite and at most 10000")
		}
		layout, err := b.textLayout(o)
		if err != nil {
			return s, err
		}
		s.width, s.height = layout.width, layout.height
		s.left, s.top = -s.width/2, -s.height/2
		s.text, s.family, s.fontSize = o.Content, layout.family, o.Size
		s.strokeWidth = 0 // The raster renderer does not stroke text either.
		s.geometry = `<a:prstGeom prst="rect"><a:avLst/></a:prstGeom>`
	case *mobject.Line:
		var p geometry.Path
		p.MoveTo(o.Start).LineTo(o.End)
		if err := s.path(p); err != nil {
			return s, err
		}
	case *mobject.Polygon:
		if len(o.Points) < 2 {
			return s, errors.New("native polygon requires at least two points")
		}
		var p geometry.Path
		p.MoveTo(o.Points[0])
		for _, point := range o.Points[1:] {
			p.LineTo(point)
		}
		p.Close()
		if err := s.path(p); err != nil {
			return s, err
		}
	case *mobject.Path:
		if err := s.path(o.Geometry); err != nil {
			return s, err
		}
	default:
		return s, fmt.Errorf("unsupported native object %T (images and 3D meshes require frame export)", obj)
	}
	for _, v := range []float64{s.left, s.top, s.width, s.height, s.strokeWidth} {
		if !nativeFinite(v) || math.Abs(v) > 1e7 {
			return s, errors.New("invalid native object dimensions")
		}
	}
	if s.width <= 0 || s.height <= 0 || s.strokeWidth < 0 {
		return s, errors.New("native object dimensions must be positive; strokeWidth must be non-negative")
	}
	return s, nil
}

func (b *nativeBuilder) textLayout(o *mobject.Text) (nativeTextLayout, error) {
	key := nativeTextKey{o.Font, o.Content, o.Size}
	if layout, ok := b.text[key]; ok {
		return layout, nil
	}
	f, err := b.loadFont(o.Font)
	if err != nil {
		return nativeTextLayout{}, err
	}
	face, err := opentype.NewFace(f.font, &opentype.FaceOptions{Size: o.Size, DPI: 72, Hinting: font.HintingNone})
	if err != nil {
		return nativeTextLayout{}, err
	}
	defer face.Close()
	lines := strings.Split(strings.ReplaceAll(o.Content, "\r\n", "\n"), "\n")
	layout := nativeTextLayout{width: o.Size, height: o.Size * 1.5 * float64(len(lines)), family: f.family}
	for _, line := range lines {
		layout.width = math.Max(layout.width, float64(font.MeasureString(face, line))/64+o.Size*.15)
	}
	b.text[key] = layout
	return layout, nil
}

func (b *nativeBuilder) loadFont(path string) (nativeFont, error) {
	if f, ok := b.fonts[path]; ok {
		return f, nil
	}
	data := fontasset.SourceHanSansSCRegular()
	if path != "" {
		var err error
		data, err = os.ReadFile(path)
		if err != nil {
			return nativeFont{}, fmt.Errorf("read native text font: %w", err)
		}
	}
	parsed, err := opentype.Parse(data)
	if err != nil {
		return nativeFont{}, fmt.Errorf("parse native text font: %w", err)
	}
	family, err := parsed.Name(nil, sfnt.NameIDFamily)
	if err != nil || family == "" {
		return nativeFont{}, errors.New("cannot determine native text font family")
	}
	f := nativeFont{parsed, family}
	b.fonts[path] = f
	return f, nil
}

func (s *nativeSpec) path(p geometry.Path) error {
	minX, minY, maxX, maxY := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	points := 0
	for _, command := range p.Commands {
		for _, point := range command.Points {
			if !nativeFinite(point.X) || !nativeFinite(point.Y) {
				return errors.New("non-finite native path point")
			}
			minX, minY, maxX, maxY = math.Min(minX, point.X), math.Min(minY, point.Y), math.Max(maxX, point.X), math.Max(maxY, point.Y)
			points++
		}
	}
	if points == 0 || points > 20000 {
		return errors.New("native paths require between 1 and 20000 control points")
	}
	s.left, s.top, s.width, s.height = minX, minY, math.Max(1, maxX-minX), math.Max(1, maxY-minY)
	var body bytes.Buffer
	pt := func(p geometry.Vec2) {
		fmt.Fprintf(&body, `<a:pt x="%d" y="%d"/>`, int64(math.Round((p.X-minX)/s.width*1000000)), int64(math.Round((p.Y-minY)/s.height*1000000)))
	}
	for _, command := range p.Commands {
		var name string
		var count int
		switch command.Type {
		case geometry.MoveCommand:
			name, count = "moveTo", 1
		case geometry.LineCommand:
			name, count = "lnTo", 1
		case geometry.QuadraticCommand:
			name, count = "quadBezTo", 2
		case geometry.CubicCommand:
			name, count = "cubicBezTo", 3
		case geometry.CloseCommand:
			body.WriteString(`<a:close/>`)
			continue
		default:
			return errors.New("unsupported native path command")
		}
		if len(command.Points) != count {
			return errors.New("invalid native path command point count")
		}
		body.WriteString("<a:" + name + ">")
		for _, p := range command.Points {
			pt(p)
		}
		body.WriteString("</a:" + name + ">")
	}
	s.geometry = `<a:custGeom><a:avLst/><a:gdLst/><a:ahLst/><a:cxnLst/><a:rect l="0" t="0" r="r" b="b"/><a:pathLst><a:path w="1000000" h="1000000">` + body.String() + `</a:path></a:pathLst></a:custGeom>`
	return nil
}

func nativeFill(c color.RGBA, opacity float64) string {
	alpha := int(math.Round(float64(c.A) / 255 * opacity * 100000))
	if alpha <= 0 {
		return `<a:noFill/>`
	}
	return fmt.Sprintf(`<a:solidFill><a:srgbClr val="%02X%02X%02X"><a:alpha val="%d"/></a:srgbClr></a:solidFill>`, c.R, c.G, c.B, alpha)
}

func (d *nativeDeck) shapeXML(index int, peak float64) string {
	track := d.tracks[index]
	s, state := track.spec, track.states[0]
	xRatio, yRatio := float64(d.width)/d.canvasWidth(), float64(d.height)/d.canvasHeight()
	x, y := int64(math.Round((state.x-state.w/2)*xRatio)), int64(math.Round((state.y-state.h/2)*yRatio))
	w, h := int64(math.Round(state.w*xRatio)), int64(math.Round(state.h*yRatio))
	var out bytes.Buffer
	fmt.Fprintf(&out, `<p:sp><p:nvSpPr><p:cNvPr id="%d" name="Object %d"/><p:cNvSpPr/><p:nvPr/></p:nvSpPr><p:spPr><a:xfrm rot="%d"><a:off x="%d" y="%d"/><a:ext cx="%d" cy="%d"/></a:xfrm>%s`, index+2, index+1, int64(math.Round(state.rotation*60000)), x, y, w, h, s.geometry)
	if s.family != "" {
		out.WriteString(`<a:noFill/>`)
	} else {
		out.WriteString(nativeFill(s.fill, peak))
	}
	fmt.Fprintf(&out, `<a:ln w="%d">`, int64(math.Round(s.strokeWidth*xRatio*state.w/s.width)))
	if s.strokeWidth == 0 {
		out.WriteString(`<a:noFill/>`)
	} else {
		out.WriteString(nativeFill(s.stroke, peak))
	}
	out.WriteString(`</a:ln></p:spPr>`)
	if s.family != "" {
		// Native fonts are referenced by family, not embedded. Keep text editable.
		size := int(math.Round(s.fontSize * state.w / s.width * xRatio / 914400 * 72 * 100))
		out.WriteString(`<p:txBody><a:bodyPr wrap="none" lIns="0" tIns="0" rIns="0" bIns="0" anchor="ctr"><a:noAutofit/></a:bodyPr><a:lstStyle/>`)
		for _, line := range strings.Split(strings.ReplaceAll(s.text, "\r\n", "\n"), "\n") {
			fmt.Fprintf(&out, `<a:p><a:pPr algn="ctr"/><a:r><a:rPr lang="zh-CN" sz="%d">%s<a:latin typeface="%s"/><a:ea typeface="%s"/></a:rPr><a:t>%s</a:t></a:r><a:endParaRPr sz="%d"/></a:p>`, size, nativeFill(s.fill, peak), pptxEscape(s.family), pptxEscape(s.family), pptxEscape(line), size)
		}
		out.WriteString(`</p:txBody>`)
	}
	out.WriteString(`</p:sp>`)
	return out.String()
}
