package gg

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/image/font/gofont/goregular"

	"github.com/23dd/animation/mobject"
	"github.com/23dd/animation/scene"
)

func TestUnstrokedBackgroundKeepsItsColor(t *testing.T) {
	for _, strokeWidth := range []float64{0, 2} {
		s := scene.New(100, 100)
		background := mobject.NewRectangle(100, 100)
		green := color.RGBA{R: 0x12, G: 0x56, B: 0x37, A: 255}
		background.Fill = green
		background.Stroke = color.RGBA{}
		background.StrokeWidth = strokeWidth
		ball := mobject.NewCircle(10)
		ball.Fill = color.RGBA{B: 255, A: 255}
		s.Add(background, ball)
		r := New(100, 100)
		for _, opacity := range []float64{1, 0.5, 0, 1} {
			ball.Opacity = opacity
			img := r.Render(s)
			got := color.RGBAModel.Convert(img.At(10, 10)).(color.RGBA)
			if got != green {
				t.Fatalf("strokeWidth=%v opacity=%v: background=%v, want %v", strokeWidth, opacity, got, green)
			}
		}
	}
}

func TestRenderCustomTrueTypeFont(t *testing.T) {
	dir := t.TempDir()
	fontPath := filepath.Join(dir, "custom.ttf")
	if err := os.WriteFile(fontPath, goregular.TTF, 0600); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(fontPath)
	if err != nil {
		t.Fatal(err)
	}
	label := mobject.NewText("TTF", 36)
	label.Font = fontPath
	label.FontVersion = info.ModTime().UnixNano() ^ info.Size()
	s := scene.New(320, 180)
	s.Add(label)
	img := New(320, 180).Render(s)
	background := s.Camera.Background
	foundTextPixel := false
	for y := 0; y < 180 && !foundTextPixel; y++ {
		for x := 0; x < 320; x++ {
			if color.RGBAModel.Convert(img.At(x, y)).(color.RGBA) != background {
				foundTextPixel = true
				break
			}
		}
	}
	if !foundTextPixel {
		t.Fatal("custom font rendered no visible pixels")
	}
}

func TestRenderEmbeddedFontWithChinese(t *testing.T) {
	label := mobject.NewText("中文动画播放", 36)
	s := scene.New(320, 180)
	s.Add(label)
	img := New(320, 180).Render(s)
	background := s.Camera.Background
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			pixel := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)
			if pixel != background {
				return
			}
		}
	}
	t.Fatal("embedded font produced no visible Chinese text")
}

func TestDefaultFontSelectsJetBrainsMonoThenChineseFallback(t *testing.T) {
	r := New(320, 180)
	face, ok := r.fontFace("", 0, 24).(*fallbackFace)
	if !ok {
		t.Fatalf("default face type: %T", r.fontFace("", 0, 24))
	}
	if face.faceFor('A') != face.primary {
		t.Fatal("Latin character did not select JetBrains Mono")
	}
	if supportsRune(face.primary, '中') {
		t.Fatal("JetBrains Mono unexpectedly contains a Chinese glyph")
	}
	if !supportsRune(face.fallback, '中') {
		t.Fatal("Source Han Sans fallback is missing a Chinese glyph")
	}
	if face.faceFor('中') != face.fallback {
		t.Fatal("Chinese character did not select Source Han Sans fallback")
	}
}

func TestRectangleDrawsProgressively(t *testing.T) {
	s := scene.New(100, 100)
	rect := mobject.NewRectangle(100, 100)
	rect.Fill = color.RGBA{R: 255, A: 255}
	rect.Stroke = color.RGBA{}
	rect.StrokeWidth = 4
	s.Add(rect)
	r := New(100, 100)
	bg := s.Camera.Background

	// Half drawn: the perimeter is only traced, so the center stays empty.
	rect.DrawProgress = 0.5
	img := r.Render(s)
	if got := color.RGBAModel.Convert(img.At(50, 50)).(color.RGBA); got != bg {
		t.Fatalf("center unfilled at progress 0.5: %v", got)
	}
	stroked := false
	for y := 0; y < 100 && !stroked; y++ {
		for x := 0; x < 100; x++ {
			if color.RGBAModel.Convert(img.At(x, y)).(color.RGBA) != bg {
				stroked = true
				break
			}
		}
	}
	if !stroked {
		t.Fatal("no stroke pixels at progress 0.5")
	}

	// Fully drawn: the rectangle is filled as usual.
	rect.DrawProgress = 1
	img = r.Render(s)
	if got := color.RGBAModel.Convert(img.At(50, 50)).(color.RGBA); got == bg {
		t.Fatal("center filled at progress 1")
	}
}

func TestTextRevealFadesInCharacter(t *testing.T) {
	label := mobject.NewText("M", 64)
	label.Fill = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	s := scene.New(100, 100)
	s.Add(label)
	r := New(100, 100)

	brightest := func(img image.Image) uint8 {
		var max uint8
		for y := 0; y < img.Bounds().Dy(); y++ {
			for x := 0; x < img.Bounds().Dx(); x++ {
				if c := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA); c.R > max {
					max = c.R
				}
			}
		}
		return max
	}

	// Render returns the renderer's internal buffer, so measure each frame
	// before the next render overwrites it.
	label.Reveal = 0.5
	halfMax := brightest(r.Render(s))
	label.Reveal = 1
	fullMax := brightest(r.Render(s))

	if fullMax < 200 {
		t.Fatalf("full reveal should be bright, max=%d", fullMax)
	}
	if halfMax >= fullMax {
		t.Fatalf("half reveal should be dimmer than full: half=%d full=%d", halfMax, fullMax)
	}
}

func TestTextCaretRendersWhenVisible(t *testing.T) {
	label := mobject.NewText("M", 64)
	label.Fill = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	label.Cursor = color.RGBA{R: 255, A: 255} // red caret
	label.Reveal = 0.5
	s := scene.New(200, 100)
	s.Add(label)
	r := New(200, 100)

	hasRed := func(img image.Image) bool {
		for y := 0; y < img.Bounds().Dy(); y++ {
			for x := 0; x < img.Bounds().Dx(); x++ {
				c := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)
				if c.R > 200 && c.G < 100 && c.B < 100 {
					return true
				}
			}
		}
		return false
	}

	label.CursorVisible = true
	if !hasRed(r.Render(s)) {
		t.Fatal("expected a red caret pixel when visible")
	}
	label.CursorVisible = false
	if hasRed(r.Render(s)) {
		t.Fatal("expected no caret pixel when hidden")
	}
}

func TestOutputSizePreservesLogicalCanvasSize(t *testing.T) {
	r := New(1280, 720)
	r.SetOutputSize(800, 450)
	w, h := r.Size()
	if w != 1280 || h != 720 {
		t.Fatalf("logical size = %dx%d", w, h)
	}
	label := mobject.NewText("Sharp 清晰", 13)
	s := scene.New(1280, 720)
	s.Add(label)
	img := r.Render(s)
	if got := img.Bounds().Size(); got.X != 800 || got.Y != 450 {
		t.Fatalf("output size = %dx%d", got.X, got.Y)
	}
}
