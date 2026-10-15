package export

import (
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/23dd/animation/scene"
	"github.com/23dd/animation/timeline"
)

type gifTestRenderer struct{ frame int }

func (r *gifTestRenderer) Size() (int, int) { return 24, 16 }
func (r *gifTestRenderer) Render(*scene.Scene) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 24, 16))
	c := color.RGBA{R: 255, A: 255}
	if r.frame%2 != 0 {
		c = color.RGBA{B: 255, A: 255}
	}
	draw.Draw(img, img.Bounds(), &image.Uniform{C: c}, image.Point{}, draw.Src)
	r.frame++
	return img
}

func TestGIFStreamFramesAndTiming(t *testing.T) {
	for _, fps := range []float64{30, 60} {
		path := filepath.Join(t.TempDir(), "stream.gif")
		tl := timeline.New()
		tl.Wait(1)
		last := 0
		err := GIFWithProgress(path, scene.New(24, 16), tl, &gifTestRenderer{}, fps, func(done, total int) {
			if done != last+1 || total != int(fps)+1 {
				t.Fatalf("progress %d/%d after %d", done, total, last)
			}
			last = done
		})
		if err != nil {
			t.Fatal(err)
		}
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		out, err := gif.DecodeAll(f)
		f.Close()
		if err != nil {
			t.Fatal(err)
		}
		if len(out.Image) != int(fps)+1 || out.LoopCount != 0 {
			t.Fatalf("invalid animation: frames=%d loop=%d", len(out.Image), out.LoopCount)
		}
		totalDelay := 0
		for i, img := range out.Image {
			if img.Bounds() != image.Rect(0, 0, 24, 16) {
				t.Fatal(img.Bounds())
			}
			r, _, b, _ := img.At(12, 8).RGBA()
			if (i%2 == 0 && r <= b) || (i%2 == 1 && b <= r) {
				t.Fatalf("frame %d lost its colors", i)
			}
			totalDelay += out.Delay[i]
		}
		if totalDelay != int(math.Round(float64(len(out.Image))*100/fps)) {
			t.Fatalf("timing drift: %d", totalDelay)
		}
	}
}

func BenchmarkGIFQuantization(b *testing.B) {
	src := image.NewRGBA(image.Rect(0, 0, 640, 360))
	for y := 0; y < 360; y++ {
		for x := 0; x < 640; x++ {
			src.SetRGBA(x, y, color.RGBA{uint8(x * 255 / 640), uint8(y * 255 / 360), 90, 255})
		}
	}
	dst := image.NewPaletted(src.Bounds(), palette.Plan9)
	lookupOnce.Do(buildLookup)
	b.Run("Old", func(b *testing.B) {
		for b.Loop() {
			draw.FloydSteinberg.Draw(dst, dst.Bounds(), src, image.Point{})
		}
	})
	b.Run("Lookup", func(b *testing.B) {
		for b.Loop() {
			quantizeGIF(dst, src)
		}
	})
}
