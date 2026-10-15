package gg

import (
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/image/font/gofont/goregular"

	"github.com/23dd/animation/mobject"
	"github.com/23dd/animation/scene"
)

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
