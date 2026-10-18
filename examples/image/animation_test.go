package imageexample_test

import (
	"crypto/sha256"
	"image"
	"math"
	"testing"

	"github.com/23dd/animation/mobject"
	ggrenderer "github.com/23dd/animation/renderer/gg"
	"github.com/23dd/animation/script"
)

func TestImageExampleFramesAndBackwardSeeking(t *testing.T) {
	r, err := script.LoadFile("animation.js", 1280, 720)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(r.Timeline.Duration()-16) > 1e-9 {
		t.Fatalf("unexpected example duration: %g", r.Timeline.Duration())
	}
	renderer := ggrenderer.New(1280, 720)
	if err := renderer.Err(); err != nil {
		t.Fatal(err)
	}
	frame := func(at float64) [32]byte {
		r.Timeline.Seek(at)
		if err := r.Scene.Update(at, 0); err != nil {
			t.Fatal(err)
		}
		return sha256.Sum256(renderer.Render(r.Scene).(*image.RGBA).Pix)
	}
	times := []float64{0, .5, 4, 6.4, 8, 12.55, 13.65, 16}
	hashes := make(map[float64][32]byte)
	for _, at := range times {
		hashes[at] = frame(at)
		if at == 12.55 || at == 13.65 {
			value, err := r.VM.RunString(`cards[1].photo.__native`)
			if err != nil {
				t.Fatal(err)
			}
			pic := value.Export().(*mobject.Image)
			if !pic.IsVisible || math.Abs(pic.DrawProgress-.5) > 1e-8 {
				t.Fatalf("create/uncreate reveal at %g: visible=%v progress=%g", at, pic.IsVisible, pic.DrawProgress)
			}
		}
	}
	for i := len(times) - 1; i >= 0; i-- {
		at := times[i]
		if hashes[at] != frame(at) {
			t.Fatalf("seeking backward changed the image frame at %g", at)
		}
	}
}
