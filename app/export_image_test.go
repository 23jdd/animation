package app

import (
	"image"
	"image/color"
	"image/gif"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestBitmapImagePNGAndGIFExport(t *testing.T) {
	dir := t.TempDir()
	pixels := image.NewNRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			if x < 5 {
				pixels.SetNRGBA(x, y, color.NRGBA{R: 255, A: 255})
			}
		}
	}
	f, err := os.Create(filepath.Join(dir, "asset.png"))
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, pixels); err != nil {
		f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(dir, "scene.js")
	if err := os.WriteFile(source, []byte(`add(image("asset.png",{width:20}));wait(.5);`), 0600); err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"png", "gif"} {
		t.Run(kind, func(t *testing.T) {
			output := filepath.Join(dir, "frame."+kind)
			if err := runRender([]string{"-width", "80", "-height", "40", "-scale", "2", "-fps", "2", source, output}, kind); err != nil {
				t.Fatal(err)
			}
			f, err := os.Open(output)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			var frame image.Image
			if kind == "png" {
				frame, err = png.Decode(f)
			} else {
				frame, err = gif.Decode(f)
			}
			if err != nil {
				t.Fatal(err)
			}
			if frame.Bounds().Dx() != 160 || frame.Bounds().Dy() != 80 {
				t.Fatalf("unexpected scaled image bounds: %v", frame.Bounds())
			}
			r, _, _, _ := frame.At(70, 40).RGBA()
			if r < 60000 {
				t.Fatal("bitmap was not rendered at its scaled position")
			}
			r, _, _, _ = frame.At(90, 40).RGBA()
			if r > 15000 {
				t.Fatal("transparent pixels did not preserve the background")
			}
		})
	}
}
