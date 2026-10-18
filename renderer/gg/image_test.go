package gg

import (
	"image"
	"image/color"
	stddraw "image/draw"
	"math"
	"testing"

	"github.com/23dd/animation/geometry"
	"github.com/23dd/animation/mobject"
	"github.com/23dd/animation/scene"
)

func solidBitmap(bounds image.Rectangle, c color.Color) *image.RGBA {
	pixels := image.NewRGBA(bounds)
	stddraw.Draw(pixels, bounds, image.NewUniform(c), image.Point{}, stddraw.Src)
	return pixels
}

func TestImageFitModesAndNonzeroSourceBounds(t *testing.T) {
	for _, fit := range []mobject.ImageFit{mobject.ImageContain, mobject.ImageCover, mobject.ImageStretch} {
		t.Run(string(fit), func(t *testing.T) {
			s := scene.New(100, 100)
			s.Camera.Background = color.RGBA{A: 255}
			pic := mobject.NewImage(solidBitmap(image.Rect(10, 20, 30, 30), color.RGBA{R: 255, A: 255}))
			pic.Width, pic.Height, pic.Fit = 40, 40, fit
			s.Add(pic)
			pixels := New(100, 100).Render(s).(*image.RGBA)
			if pixels.RGBAAt(50, 50).R < 250 {
				t.Fatal("image is not centered or source bounds were not normalized")
			}
			if got := pixels.RGBAAt(50, 35).R; (fit == mobject.ImageContain && got != 0) || (fit != mobject.ImageContain && got < 250) {
				t.Fatalf("incorrect vertical fit: red=%d", got)
			}
			if pixels.RGBAAt(25, 50).R != 0 || pixels.RGBAAt(50, 25).R != 0 {
				t.Fatal("image overflowed its display box")
			}
		})
	}
}

func TestImageAlphaGroupCameraAndExportScale(t *testing.T) {
	s := scene.New(100, 100)
	s.Camera.Background = color.RGBA{A: 255}
	s.Camera.Position.X, s.Camera.Zoom = 4, 2
	pixels := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	stddraw.Draw(pixels, pixels.Bounds(), image.NewUniform(color.NRGBA{R: 255, A: 128}), image.Point{}, stddraw.Src)
	pic := mobject.NewImage(pixels)
	pic.Width, pic.Height, pic.Opacity = 10, 10, .5
	pic.SetPosition(geometry.V2(5, 0))
	g := mobject.NewGroup(pic)
	g.Opacity = .5
	g.SetPosition(geometry.V2(10, 0))
	g.Transform.Scale = geometry.V2(2, 2)
	s.Add(g)
	r := New(100, 100)
	r.SetOutputSize(200, 200)
	frame := r.Render(s).(*image.RGBA)
	if got := frame.RGBAAt(164, 100); got.R < 30 || got.R > 34 || got.A != 255 {
		t.Fatalf("source/group/object alpha or camera transform incorrect: %v", got)
	}
	if frame.RGBAAt(115, 100).R != 0 {
		t.Fatal("output scaling changed the logical image footprint")
	}
	if got := pixels.NRGBAAt(2, 2); got != (color.NRGBA{R: 255, A: 128}) {
		t.Fatalf("rendering mutated source pixels: %v", got)
	}
}

func TestImageRotationMirroringAndCoverClip(t *testing.T) {
	s := scene.New(100, 100)
	s.Camera.Background = color.RGBA{A: 255}
	pixels := image.NewRGBA(image.Rect(0, 0, 20, 10))
	stddraw.Draw(pixels, image.Rect(0, 0, 10, 10), image.NewUniform(color.RGBA{R: 255, A: 255}), image.Point{}, stddraw.Src)
	stddraw.Draw(pixels, image.Rect(10, 0, 20, 10), image.NewUniform(color.RGBA{B: 255, A: 255}), image.Point{}, stddraw.Src)
	pic := mobject.NewImage(pixels)
	pic.Width, pic.Height, pic.Fit = 40, 20, mobject.ImageCover
	pic.Transform.Scale.X, pic.Transform.Rotation = -1, math.Pi/2
	s.Add(pic)
	frame := New(100, 100).Render(s).(*image.RGBA)
	if frame.RGBAAt(50, 35).B < 240 || frame.RGBAAt(50, 65).R < 240 {
		t.Fatal("rotation/mirroring did not transform image pixels")
	}
	if frame.RGBAAt(35, 50).B != 0 || frame.RGBAAt(35, 50).R != 0 {
		t.Fatal("cover clipping did not rotate with its image")
	}
}

func TestImageRevealAndClipDoNotAffectFollowingObjects(t *testing.T) {
	s := scene.New(100, 100)
	s.Camera.Background = color.RGBA{A: 255}
	pic := mobject.NewImage(solidBitmap(image.Rect(0, 0, 20, 10), color.RGBA{R: 255, A: 255}))
	pic.Width, pic.Height, pic.DrawProgress = 40, 40, .5
	marker := mobject.NewRectangle(8, 8)
	marker.SetPosition(geometry.V2(30, 0))
	marker.Fill, marker.Stroke = color.RGBA{G: 255, A: 255}, color.RGBA{}
	s.Add(pic, marker)
	r := New(100, 100)
	frame := r.Render(s).(*image.RGBA)
	if frame.RGBAAt(40, 50).R < 240 || frame.RGBAAt(60, 50).R != 0 {
		t.Fatal("create should reveal the bitmap from left to right")
	}
	if frame.RGBAAt(80, 50).G < 240 {
		t.Fatal("image clipping leaked into following objects")
	}
	pic.DrawProgress = 0
	if r.Render(s).(*image.RGBA).RGBAAt(40, 50).R != 0 {
		t.Fatal("zero reveal still drew image pixels")
	}
	pic.DrawProgress, pic.Transform.Scale.X = 1, 0
	r.Render(s) // Degenerate transforms must not panic.
}
