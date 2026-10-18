package rosecurve_test

import (
	"crypto/sha256"
	"fmt"
	"image"
	"testing"

	ggrenderer "github.com/23dd/animation/renderer/gg"
	"github.com/23dd/animation/script"
)

func TestRoseCoordinatesAndPetalCount(t *testing.T) {
	runtime, err := script.LoadFile("animation.js", 1280, 720)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.Timeline.Duration() != 64 {
		t.Fatal("incorrect film duration")
	}
	checks := []string{
		`[1,3,5,7].every(n=>petalCount(n)===n && period(n)===Math.PI)`,
		`[2,4,6,8].every(n=>petalCount(n)===2*n && period(n)===TAU)`,
		`Math.abs(rose(5,0)[0]-1)<1e-12 && Math.abs(rose(5,0)[1])<1e-12`,
		`Math.hypot(...rose(5,startAngle(5)))<1e-12 && Math.hypot(...rose(5,startAngle(5)+Math.PI))<1e-12`,
		`(()=>{const t=Math.PI/5,p=rose(5,t);return Math.abs(p[0]+Math.cos(t))<1e-12 && Math.abs(p[1]+Math.sin(t))<1e-12;})()`,
		`drawingProgress(14.5)===0 && drawingProgress(25.5)===1`,
	}
	for _, check := range checks {
		v, err := runtime.VM.RunString(check)
		if err != nil || !v.ToBoolean() {
			t.Fatalf("failed %s: %v %v", check, v, err)
		}
	}
	for i := 0; i < 65; i++ {
		value, err := runtime.VM.RunString(fmt.Sprintf(`(()=>{
const t=%.17g,a=rose(5,t),b=rose(5,t+TAU/5),c=Math.cos(TAU/5),s=Math.sin(TAU/5);
return Math.hypot(...a)<=1.000000001 && Math.abs(b[0]-(a[0]*c-a[1]*s))<1e-10 && Math.abs(b[1]-(a[0]*s+a[1]*c))<1e-10;
})()`, float64(i)*.137))
		if err != nil || !value.ToBoolean() {
			t.Fatalf("fivefold symmetry sample %d: %v %v", i, value, err)
		}
	}
}

func TestTraceHeadAndCompletion(t *testing.T) {
	runtime, err := script.LoadFile("animation.js", 1280, 720)
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range []int{3, 4, 5, 7} {
		if _, err := runtime.VM.RunString(fmt.Sprintf(`globalThis.probe=makeTracer(%d,151);`, n)); err != nil {
			t.Fatal(err)
		}
		for _, progress := range []float64{.001, .02, .123, .5, .876, .999} {
			value, err := runtime.VM.RunString(fmt.Sprintf(`(()=>{
const p=%.17g;probe.setProgress(p);
const j=Math.floor(p*FILM.samples)%%32,segment=probe.partial[j];
const expected=screenRose(%d,startAngle(%d)+period(%d)*p,151);
return Math.max(Math.abs(segment.x+segment.scaleX*Math.cos(segment.rotation)-expected[0]),Math.abs(segment.y+segment.scaleX*Math.sin(segment.rotation)-expected[1]));
})()`, progress, n, n, n))
			if err != nil || value.ToFloat() > 1e-8 {
				t.Fatalf("n=%d trace head mismatch at %g: %v %v", n, progress, value, err)
			}
		}
		value, err := runtime.VM.RunString(`probe.setProgress(1);probe.blocks.every(b=>b.opacity===1)&&probe.partial.every(p=>p.opacity===0)&&probe.fills.every(p=>p.opacity===.5)&&Math.hypot(...probe.head)<1e-8`)
		if err != nil || !value.ToBoolean() {
			t.Fatalf("n=%d completed trace did not close cleanly", n)
		}
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
	times := []float64{0, 3, 7, 11, 14, 14.5, 20, 25.5, 27, 33, 36, 39, 43, 45, 50, 54, 56, 60, 63.8, 64}
	before := make(map[float64][32]byte)
	for _, at := range times {
		before[at] = frame(at)
	}
	for i := len(times) - 1; i >= 0; i-- {
		at := times[i]
		if before[at] != frame(at) {
			t.Fatalf("backward seeking changed the frame at %g", at)
		}
	}
}

func TestEnrichedScenesAndMotion(t *testing.T) {
	runtime, err := script.LoadFile("animation.js", 1280, 720)
	if err != nil {
		t.Fatal(err)
	}
	for _, check := range []string{
		`SCENES.length===7 && bouquetFlowers.length===7 && wreathItems.length===12 && bouquetBuds.length===3`,
		`galleryItems.map(o=>o.n).join(",")==="3,5,7" && spiralPetals.length===30`,
		`(()=>{for(let i=0;i<30;i++)for(const t of [35,36,38,39,40,43,45]){const s=spiralState(i,t);if(![s.x,s.y,s.rotation,s.opacity,s.scale].every(Number.isFinite)||s.opacity<0||s.opacity>1||Math.hypot(s.x,(s.y+35)/.68)>285.000001)return false;}return true;})()`,
		`spiralState(0,35).opacity===0 && spiralState(0,39).opacity>.1 && spiralState(0,45).opacity<1e-12`,
	} {
		v, err := runtime.VM.RunString(check)
		if err != nil || !v.ToBoolean() {
			t.Fatalf("composition check failed: %s: %v %v", check, v, err)
		}
	}
	for i, at := range []float64{3, 11, 21, 33, 41, 50, 60} {
		runtime.Timeline.Seek(at)
		if err := runtime.Scene.Update(at, 0); err != nil {
			t.Fatal(err)
		}
		v, err := runtime.VM.RunString(fmt.Sprintf(`SCENES.every((s,i)=>i===%d?s.g.opacity===1:s.g.opacity===0)`, i))
		if err != nil || !v.ToBoolean() {
			t.Fatalf("wrong scene visible at %g: %v %v", at, v, err)
		}
	}
	for _, at := range []float64{55, 60} {
		runtime.Timeline.Seek(at)
		if err := runtime.Scene.Update(at, 0); err != nil {
			t.Fatal(err)
		}
		check := `bouquetFlowers[6].g.opacity<.1 && bow.opacity===0`
		if at == 60 {
			check = `bouquetFlowers.every(f=>f.g.opacity===1) && bow.opacity===1 && giftTag.opacity===1`
		}
		v, err := runtime.VM.RunString(check)
		if err != nil || !v.ToBoolean() {
			t.Fatalf("bouquet did not assemble progressively at %g: %v %v", at, v, err)
		}
	}
}
