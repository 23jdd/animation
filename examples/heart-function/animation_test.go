package heartfunction_test

import (
	"crypto/sha256"
	"fmt"
	"image"
	"math"
	"testing"

	ggrenderer "github.com/23dd/animation/renderer/gg"
	"github.com/23dd/animation/script"
)

func TestHeartCoordinatesAndSymmetry(t *testing.T) {
	runtime, err := script.LoadFile("animation.js", 1280, 720)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.Timeline.Duration() != 44 {
		t.Fatal("incorrect duration")
	}
	for _, p := range []struct{ t, x, y float64 }{{0, 0, 5}, {math.Pi / 2, 16, 4}, {math.Pi, 0, -17}, {3 * math.Pi / 2, -16, 4}, {2 * math.Pi, 0, 5}} {
		v, err := runtime.VM.RunString(fmt.Sprintf(`Math.max(Math.abs(heart(%.17g)[0]-(%.17g)),Math.abs(heart(%.17g)[1]-(%.17g)))`, p.t, p.x, p.t, p.y))
		if err != nil || v.ToFloat() > 1e-10 {
			t.Fatalf("wrong key point %+v: %v %v", p, v, err)
		}
	}
	for i := 0; i < 65; i++ {
		at := float64(i) * 2 * math.Pi / 64
		v, err := runtime.VM.RunString(fmt.Sprintf(`(()=>{const a=heart(%.17g),b=heart(TAU-%.17g);return Math.max(Math.abs(a[0]+b[0]),Math.abs(a[1]-b[1]));})()`, at, at))
		if err != nil || v.ToFloat() > 1e-10 {
			t.Fatalf("symmetry at %g: %v %v", at, v, err)
		}
	}
	v, err := runtime.VM.RunString(`(()=>{const s=fitScale(300,240);return (BOUNDS.maxX-BOUNDS.minX)*s<=300.000001 && (BOUNDS.maxY-BOUNDS.minY)*s<=240.000001;})()`)
	if err != nil || !v.ToBoolean() {
		t.Fatal("auto-fit exceeded its viewport")
	}
}

func TestTraceHeadMatchesParameterNotArcLength(t *testing.T) {
	runtime, err := script.LoadFile("animation.js", 1280, 720)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runtime.VM.RunString(`globalThis.probeTrace=makeTracer(9);`); err != nil {
		t.Fatal(err)
	}
	for _, at := range []float64{.001, .123, .876, math.Pi / 2, math.Pi, 1.9 * math.Pi, 2*math.Pi - .001} {
		value, err := runtime.VM.RunString(fmt.Sprintf(`(()=>{
const t=%.17g;probeTrace.setParameter(t);
const index=Math.floor(t/TAU*SAMPLES),j=index%%32;
const line=probeTrace.partial[j],head=screenPoint(t,9);
return Math.max(Math.abs(line.x+line.scaleX*Math.cos(line.rotation)-head[0]),Math.abs(line.y+line.scaleX*Math.sin(line.rotation)-head[1]));
})()`, at))
		if err != nil || value.ToFloat() > 1e-8 {
			t.Fatalf("trace head mismatch at %g: %v %v", at, value, err)
		}
	}
	value, err := runtime.VM.RunString(`probeTrace.setParameter(TAU);probeTrace.blocks.every(b=>b.opacity===1)&&probeTrace.partial.every(p=>p.opacity===0)`)
	if err != nil || !value.ToBoolean() {
		t.Fatal("completed trace did not close cleanly")
	}
}

func TestFilmFramesAndBackwardSeeking(t *testing.T) {
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
	before := frame(18)
	for _, at := range []float64{0, 2, 8, 12.5, 16.5, 20.5, 24.5, 27.5, 31, 36, 40, 43.8, 44, 2} {
		frame(at)
	}
	if before != frame(18) {
		t.Fatal("backward seeking changed the frame")
	}
	frame(20.5)
	v, err := runtime.VM.RunString(`Math.abs(tracer.t-Math.PI)<1e-10 && Math.abs(point.x)<1e-8 && Math.abs(point.y-17*plotScale)<1e-8`)
	if err != nil || !v.ToBoolean() {
		t.Fatal("the pause at the bottom tip is incorrect")
	}
}
