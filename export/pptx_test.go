package export

import (
	"archive/zip"
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/23dd/animation/mobject"
	"github.com/23dd/animation/scene"
	"github.com/23dd/animation/timeline"
)

type pptxTestRenderer struct {
	object *mobject.Circle
	pixels *image.RGBA
	calls  int
}

func (r *pptxTestRenderer) Size() (int, int) { return 32, 18 }
func (r *pptxTestRenderer) Render(*scene.Scene) image.Image {
	if r.pixels == nil {
		r.pixels = image.NewRGBA(image.Rect(0, 0, 64, 36))
	}
	r.calls++
	red := uint8(r.object.Position().X)
	draw.Draw(r.pixels, r.pixels.Bounds(), image.NewUniform(color.RGBA{R: red, A: 255}), image.Point{}, draw.Src)
	return r.pixels
}

func pptxTestScene() (*scene.Scene, *timeline.Timeline, *pptxTestRenderer) {
	s, tl := scene.New(32, 18), timeline.New()
	tl.Wait(2)
	obj := mobject.NewCircle(2)
	obj.AddUpdater(func(_ mobject.Mobject, at, dt float64) error {
		obj.Transform.Position.X = at * 100
		return nil
	})
	s.Add(obj)
	tl.Seek(.25)
	_ = s.Update(.25, 0)
	return s, tl, &pptxTestRenderer{object: obj}
}

func TestPPTXTimes(t *testing.T) {
	for _, times := range [][]float64{{0, 2}, {2, 0, 2}, nil} {
		got, err := validatePPTXTimes(times, 2)
		if err != nil || len(got) == 0 {
			t.Fatalf("valid times %v: %v %v", times, got, err)
		}
		if len(times) > 0 {
			got[0] = .1
			if times[0] == .1 {
				t.Fatal("frame selection aliases caller's slice")
			}
		}
	}
	for _, times := range [][]float64{{-1}, {2.1}, {math.NaN()}, {math.Inf(1)}, make([]float64, MaxPPTXSlides+1)} {
		if _, err := validatePPTXTimes(times, 2); err == nil {
			t.Fatalf("accepted invalid frame selection: %v", times)
		}
	}
	if _, err := validatePPTXTimes(nil, math.Inf(1)); err == nil {
		t.Fatal("accepted invalid duration")
	}
}

func TestPPTXSlideSizePreservesAspectRatio(t *testing.T) {
	for _, p := range []struct{ w, h, sw, sh int }{
		{1280, 720, 12192000, 6858000}, {720, 1280, 6858000, 12192000},
		{500, 500, 12192000, 12192000}, {4000, 200, 18288000, 914400},
	} {
		sw, sh, err := pptxSlideSize(p.w, p.h)
		if err != nil || sw != int32(p.sw) || sh != int32(p.sh) {
			t.Fatalf("%dx%d produced %dx%d: %v", p.w, p.h, sw, sh, err)
		}
	}
	for _, dims := range [][2]int{{0, 100}, {100, -1}, {10000, 1}} {
		if _, _, err := pptxSlideSize(dims[0], dims[1]); err == nil {
			t.Fatalf("accepted unsupported slide dimensions: %v", dims)
		}
	}
}

func TestPPTXFramePipelineOwnsPNGsAndRestoresScene(t *testing.T) {
	s, tl, r := pptxTestScene()
	var images [][]byte
	var timestamps []float64
	progressCalls := 0
	err := renderPPTXFrames(s, tl, r, []float64{2, 0, 1}, func(index int, at float64, data []byte) error {
		if index != len(images) {
			t.Fatal("slide order changed")
		}
		images = append(images, data)
		timestamps = append(timestamps, at)
		return nil
	}, func(done, total int) {
		progressCalls++
		if done != progressCalls || total != 3 {
			t.Fatal("incorrect progress")
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	for i, data := range images {
		img, err := png.Decode(bytes.NewReader(data))
		if err != nil || img.Bounds() != image.Rect(0, 0, 64, 36) {
			t.Fatalf("invalid scaled PNG for slide %d: %v", i, err)
		}
		got, _, _, _ := img.At(20, 20).RGBA()
		if got != uint32(timestamps[i]*100)*257 {
			t.Fatalf("slide %d lost its own PNG data: red=%d", i, got)
		}
	}
	if tl.Time != .25 || r.object.Position().X != 25 || progressCalls != 3 {
		t.Fatal("export did not restore the original scene/time")
	}
}

func TestPPTXFrameFailureRestoresScene(t *testing.T) {
	s, tl, r := pptxTestScene()
	failure := errors.New("test slide failure")
	err := renderPPTXFrames(s, tl, r, []float64{0, 1}, func(int, float64, []byte) error { return failure }, nil)
	if !errors.Is(err, failure) || tl.Time != .25 || r.object.Position().X != 25 {
		t.Fatalf("failed export changed original state: %v time=%g", err, tl.Time)
	}
}

func TestPPTXAtomicSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "deck.pptx")
	if err := os.WriteFile(path, []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	failure := errors.New("simulated writer error")
	if err := savePPTX(path, func(w io.Writer) error {
		_, _ = w.Write([]byte("partial"))
		return failure
	}); !errors.Is(err, failure) {
		t.Fatalf("writer error was lost: %v", err)
	}
	data, _ := os.ReadFile(path)
	entries, _ := os.ReadDir(dir)
	if string(data) != "original" || len(entries) != 1 {
		t.Fatal("failed save overwrote output or left a temporary file")
	}
	if err := savePPTX(path, func(w io.Writer) error { _, err := w.Write([]byte("new")); return err }); err != nil {
		t.Fatal(err)
	}
	data, _ = os.ReadFile(path)
	if string(data) != "new" {
		t.Fatal("successful save did not replace the output")
	}
}

func TestPPTXFailureDoesNotTouchOutput(t *testing.T) {
	s, tl, r := pptxTestScene()
	path := filepath.Join(t.TempDir(), "deck.pptx")
	if err := os.WriteFile(path, []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	err := PPTX(path, s, tl, r, []float64{-1})
	if err == nil {
		t.Fatal("invalid frame time was accepted")
	}
	data, _ := os.ReadFile(path)
	if string(data) != "original" || r.calls != 0 {
		t.Fatal("invalid input rendered images or changed the output")
	}
	failure := errors.New("updater failed")
	r.object.AddUpdater(func(_ mobject.Mobject, at, dt float64) error {
		if at == 1 {
			return failure
		}
		return nil
	})
	if err := PPTX(path, s, tl, r, []float64{0, 1}); !errors.Is(err, failure) {
		t.Fatalf("updater error was lost: %v", err)
	}
	data, _ = os.ReadFile(path)
	entries, _ := os.ReadDir(filepath.Dir(path))
	if string(data) != "original" || len(entries) != 1 || tl.Time != .25 {
		t.Fatal("failed render changed output, left a temporary file or lost scene time")
	}
}

func TestPPTXIntegration(t *testing.T) {
	s, tl, r := pptxTestScene()
	path := filepath.Join(t.TempDir(), "deck.pptx")
	if err := PPTX(path, s, tl, r, []float64{0, 1}); err != nil {
		t.Fatal(err)
	}
	z, err := zip.OpenReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	slides, media := 0, 0
	var redValues []uint32
	for _, f := range z.File {
		if strings.HasPrefix(f.Name, "ppt/slides/slide") && strings.HasSuffix(f.Name, ".xml") && !strings.Contains(f.Name, "/_rels/") {
			slides++
		}
		if strings.HasPrefix(f.Name, "ppt/media/") && strings.HasSuffix(f.Name, ".png") {
			media++
			stream, err := f.Open()
			if err != nil {
				t.Fatal(err)
			}
			img, err := png.Decode(stream)
			stream.Close()
			if err != nil || img.Bounds() != image.Rect(0, 0, 64, 36) {
				t.Fatalf("bad embedded frame: %v", err)
			}
			red, _, _, _ := img.At(20, 20).RGBA()
			redValues = append(redValues, red)
		}
	}
	if slides != 2 || media != 2 || len(redValues) != 2 || redValues[0] == redValues[1] {
		t.Fatalf("unexpected package contents: slides=%d images=%d red=%v", slides, media, redValues)
	}
}
