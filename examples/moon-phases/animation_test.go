package moonphases_test

import (
	"crypto/sha256"
	"fmt"
	"image"
	"math"
	"testing"

	"github.com/23dd/animation/mobject"
	ggrenderer "github.com/23dd/animation/renderer/gg"
	"github.com/23dd/animation/scene"
	"github.com/23dd/animation/script"
)

func TestLunarPhaseGeometry(t *testing.T) {
	runtime, err := script.LoadFile("animation.js", 1280, 720)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.Timeline.Duration() != 44 {
		t.Fatal("film duration must be 44 seconds")
	}
	for _, sample := range []struct {
		at    float64
		index int
	}{{12.5, 0}, {14, 1}, {16.5, 2}, {18, 3}, {20.5, 4}, {22, 5}, {24.5, 6}, {26, 7}, {27.5, 0}} {
		value, err := runtime.VM.RunString(fmt.Sprintf("phaseNameIndex(orbitalAngle(%g))", sample.at))
		if err != nil || value.ToInteger() != int64(sample.index) {
			t.Fatalf("incorrect phase name at %g", sample.at)
		}
	}
	probe, err := runtime.VM.RunString(`globalThis.probeMoon=moon(48,false);probeMoon.g`)
	if err != nil {
		t.Fatal(err)
	}
	native, ok := probe.ToObject(runtime.VM).Get("__native").Export().(mobject.Mobject)
	if !ok {
		t.Fatal("moon did not produce engine objects")
	}
	s := scene.New(128, 128)
	s.Add(native)
	renderer := ggrenderer.New(128, 128)
	for i := 0; i <= 8; i++ {
		angle := float64(i) * math.Pi / 4
		if _, err := runtime.VM.RunString(fmt.Sprintf("probeMoon.phase(%.17g)", angle)); err != nil {
			t.Fatal(err)
		}
		img := renderer.Render(s)
		lit, total := 0, 0
		for y := 0; y < 128; y++ {
			for x := 0; x < 128; x++ {
				if math.Hypot(float64(x)+.5-64, float64(y)+.5-64) >= 47.8 {
					continue
				}
				total++
				r, _, _, _ := img.At(x, y).RGBA()
				if r>>8 > 128 {
					lit++
				}
			}
		}
		got, want := float64(lit)/float64(total), (1-math.Cos(angle))/2
		if math.Abs(got-want) > .018 {
			t.Fatalf("phase %d: rendered lit fraction %.4f, want %.4f", i, got, want)
		}
	}
}

func TestFilmFramesAndDeterministicSeeking(t *testing.T) {
	runtime, err := script.LoadFile("animation.js", 1280, 720)
	if err != nil {
		t.Fatal(err)
	}
	renderer := ggrenderer.New(1280, 720)
	if err := renderer.Err(); err != nil {
		t.Fatal(err)
	}
	frame := func(at float64) [32]byte {
		runtime.Timeline.Seek(at)
		if err := runtime.Scene.Update(at, 0); err != nil {
			t.Fatal(err)
		}
		return sha256.Sum256(renderer.Render(runtime.Scene).(*image.RGBA).Pix)
	}
	first := frame(17)
	for _, at := range []float64{0, 2, 8, 12, 16, 20, 24, 28, 31, 38, 43.8, 44, 2} {
		frame(at)
	}
	if frame(17) != first {
		t.Fatal("backward seeking changed the scene")
	}
	value, err := runtime.VM.RunString(`photons.every(p=>p.x>=-274 && p.x<=143)`)
	if err != nil || !value.ToBoolean() {
		t.Fatal("photons left their light rays")
	}
}
