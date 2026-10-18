package script_test

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/23dd/animation/mobject"
	"github.com/23dd/animation/script"
)

func imageFixture(t *testing.T, dir, name string) string {
	t.Helper()
	pixels := image.NewNRGBA(image.Rect(0, 0, 20, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 20; x++ {
			pixels.SetNRGBA(x, y, color.NRGBA{R: 255, A: 128})
		}
	}
	var buf bytes.Buffer
	var err error
	switch filepath.Ext(name) {
	case ".jpg", ".jpeg":
		err = jpeg.Encode(&buf, pixels, nil)
	case ".gif":
		err = gif.Encode(&buf, pixels, nil)
	default:
		err = png.Encode(&buf, pixels)
	}
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, buf.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestImageRelativePathDimensionsAndCache(t *testing.T) {
	dir := t.TempDir()
	imageFixture(t, dir, "photo.png")
	scenePath := filepath.Join(dir, "scene.js")
	source := `const a=image("./photo.png");const b=image("photo.png",{width:80});const c=image("photo.png",{height:30});const d=image("photo.png",{width:40,height:40,fit:"cover"});add(a,b,c,d);`
	if err := os.WriteFile(scenePath, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	r, err := script.LoadFile(scenePath, 200, 100)
	if err != nil {
		t.Fatal(err)
	}
	objects := r.Scene.Snapshot()
	for i, want := range [][2]float64{{20, 10}, {80, 40}, {60, 30}, {40, 40}} {
		pic := objects[i].(*mobject.Image)
		if pic.Width != want[0] || pic.Height != want[1] {
			t.Fatalf("object %d dimensions=%gx%g, want %v", i, pic.Width, pic.Height, want)
		}
		if pic.Pixels != objects[0].(*mobject.Image).Pixels {
			t.Fatal("repeated image calls should share decoded pixels")
		}
	}
	value, err := r.VM.RunString(`a.naturalWidth===20 && a.naturalHeight===10 && b.width===80 && b.height===40 && d.fit==="cover"`)
	if err != nil || !value.ToBoolean() {
		t.Fatalf("image metadata incorrect: %v %v", value, err)
	}
	if _, err := r.VM.RunString(`"use strict"; a.width=300;`); err == nil {
		t.Fatal("width should be read-only; resizing uses scaleTo")
	}
	if err := os.Remove(filepath.Join(dir, "photo.png")); err != nil {
		t.Fatal(err)
	}
	if _, err := r.VM.RunString(`image("photo.png");`); err != nil {
		t.Fatalf("cached assets should not be reopened each time: %v", err)
	}
	if _, err := script.LoadFile(scenePath, 200, 100); err == nil {
		t.Fatal("a new runtime should reload assets rather than reuse a stale global cache")
	}
}

func TestImageDecodesPNGJPEGWebPAndGIF(t *testing.T) {
	dir := t.TempDir()
	r, err := script.New(100, 100)
	if err != nil {
		t.Fatal(err)
	}
	r.BaseDir = dir
	for _, name := range []string{"a.png", "b.jpg", "c.jpeg", "d.gif"} {
		imageFixture(t, dir, name)
		if _, err := r.VM.RunString(`add(image("` + name + `"));`); err != nil {
			t.Fatalf("decode %s: %v", name, err)
		}
	}
	// Original 2x2 red WebP fixture, encoded losslessly for this test.
	webp, err := base64.StdEncoding.DecodeString("UklGRhwAAABXRUJQVlA4TA8AAAAvAUAAAAcQ9Y/+ByKi/wEA")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "e.webp"), webp, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := r.VM.RunString(`add(image("e.webp"));`); err != nil {
		t.Fatalf("decode WebP: %v", err)
	}
	pic := r.Scene.Snapshot()[4].(*mobject.Image)
	if pic.Width != 2 || pic.Height != 2 {
		t.Fatalf("WebP dimensions=%gx%g", pic.Width, pic.Height)
	}
}

func TestImageGIFUsesFirstFrameWithItsLogicalCanvas(t *testing.T) {
	dir := t.TempDir()
	palette := color.Palette{color.Transparent, color.RGBA{R: 255, A: 255}, color.RGBA{B: 255, A: 255}}
	first := image.NewPaletted(image.Rect(3, 2, 9, 6), palette)
	second := image.NewPaletted(image.Rect(0, 0, 20, 10), palette)
	for i := range first.Pix {
		first.Pix[i] = 1
	}
	for i := range second.Pix {
		second.Pix[i] = 2
	}
	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, &gif.GIF{Image: []*image.Paletted{first, second}, Delay: []int{10, 10},
		Config: image.Config{ColorModel: palette, Width: 20, Height: 10}}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "animated.gif"), buf.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	r, err := script.New(100, 100)
	if err != nil {
		t.Fatal(err)
	}
	r.BaseDir = dir
	if _, err := r.VM.RunString(`add(image("animated.gif"));wait(5);`); err != nil {
		t.Fatal(err)
	}
	r.Timeline.Seek(3)
	if err := r.Scene.Update(3, 0); err != nil {
		t.Fatal(err)
	}
	pic := r.Scene.Snapshot()[0].(*mobject.Image)
	if pic.Width != 20 || pic.Height != 10 {
		t.Fatal("GIF frame should retain its logical canvas size")
	}
	red, _, blue, _ := pic.Pixels.At(5, 3).RGBA()
	if red != 65535 || blue != 0 {
		t.Fatal("GIF frame offset was lost or animation advanced past frame one")
	}
	_, _, _, alpha := pic.Pixels.At(0, 0).RGBA()
	if alpha != 0 {
		t.Fatal("pixels outside the first frame should remain transparent")
	}
}

func TestImageRejectsInvalidArgumentsAndFiles(t *testing.T) {
	dir := t.TempDir()
	imageFixture(t, dir, "photo.png")
	if err := os.WriteFile(filepath.Join(dir, "broken.png"), []byte("not an image"), 0600); err != nil {
		t.Fatal(err)
	}
	r, err := script.New(100, 100)
	if err != nil {
		t.Fatal(err)
	}
	r.BaseDir = dir
	for _, source := range []string{
		`image()`, `image(null)`, `image(123)`, `image({})`, `image("")`, `image("   ")`,
		`image("https://example.com/pic.png")`, `image("data:image/png;base64,AA==")`,
		`image("missing.png")`, `image("broken.png")`, `image(".")`,
		`image("photo.png",[])`, `image("photo.png",1)`, `image("photo.png",{fit:"wrong"})`,
		`image("photo.png",{width:0})`, `image("photo.png",{height:-1})`,
		`image("photo.png",{width:"40"})`, `image("photo.png",{width:NaN})`,
		`image("photo.png",{height:Infinity})`, `image("photo.png",{height:1e308})`,
	} {
		if _, err := r.VM.RunString(source); err == nil {
			t.Errorf("accepted invalid image call: %s", source)
		}
	}
	value, err := r.VM.RunString(`let caught=false;try{image("missing.png");}catch(e){caught=e instanceof TypeError;}caught && image("photo.png").width===20`)
	if err != nil || !value.ToBoolean() {
		t.Fatalf("image errors should be catchable without poisoning the runtime: %v %v", value, err)
	}
}

func TestImageRejectsOversizedHeaderBeforePixelDecode(t *testing.T) {
	dir := t.TempDir()
	// An otherwise valid IHDR exceeding the decoded-pixel budget. No IDAT is
	// necessary: the loader must reject the dimensions before decoding pixels.
	header := make([]byte, 13)
	binary.BigEndian.PutUint32(header[0:4], 8192)
	binary.BigEndian.PutUint32(header[4:8], 8192)
	header[8], header[9] = 8, 6
	var buf bytes.Buffer
	buf.WriteString("\x89PNG\r\n\x1a\n")
	_ = binary.Write(&buf, binary.BigEndian, uint32(13))
	buf.WriteString("IHDR")
	buf.Write(header)
	_ = binary.Write(&buf, binary.BigEndian, crc32.ChecksumIEEE(append([]byte("IHDR"), header...)))
	if err := os.WriteFile(filepath.Join(dir, "huge.png"), buf.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	r, err := script.New(100, 100)
	if err != nil {
		t.Fatal(err)
	}
	r.BaseDir = dir
	_, err = r.VM.RunString(`image("huge.png")`)
	if err == nil || !strings.Contains(err.Error(), "64 million") {
		t.Fatalf("oversized header was not rejected early: %v", err)
	}
}

func TestImageAnimationsAndBackwardSeeking(t *testing.T) {
	dir := t.TempDir()
	imageFixture(t, dir, "photo.png")
	r, err := script.New(200, 100)
	if err != nil {
		t.Fatal(err)
	}
	r.BaseDir = dir
	_, err = r.VM.RunString(`const pic=image("photo.png",{position:[-20,0],opacity:0});add(pic);play(fadeIn(pic));play(parallel(moveTo(pic,[40,10]),rotateTo(pic,Math.PI/2),scaleTo(pic,2)),{duration:2});play(fadeOut(pic));`)
	if err != nil {
		t.Fatal(err)
	}
	pic := r.Scene.Snapshot()[0].(*mobject.Image)
	for _, at := range []float64{2, 4, 0, 2} {
		r.Timeline.Seek(at)
		if at == 2 {
			if p := pic.Position(); p.X != 10 || p.Y != 5 || pic.Opacity != 1 || !pic.IsVisible || math.Abs(pic.Transform.Rotation-math.Pi/4) > 1e-10 || pic.Transform.Scale.X != 1.5 {
				t.Fatalf("image animation state at 2s: %+v", pic.Base())
			}
		}
		if at == 4 && pic.IsVisible {
			t.Fatal("fadeOut did not hide the image")
		}
	}
}
