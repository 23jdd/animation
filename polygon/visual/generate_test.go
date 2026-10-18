package visual

import (
	"crypto/sha256"
	"image"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/23dd/animation/polygon"
	ggrenderer "github.com/23dd/animation/renderer/gg"
	"github.com/23dd/animation/script"
)

func TestJavaScriptSceneAndBackwardSeeking(t *testing.T) {
	r, err := polygon.Solve(polygon.Demo())
	if err != nil {
		t.Fatal(err)
	}
	source, err := Script(r)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(source, "/*__POLYGON_DATA__*/null") {
		t.Fatal("unreplaced data marker")
	}
	file := filepath.Join(t.TempDir(), "animation.js")
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	runtime, err := script.LoadFile(file, 1280, 720)
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
		im := renderer.Render(runtime.Scene).(*image.RGBA)
		return sha256.Sum256(im.Pix)
	}
	first := frame(8)
	frame(runtime.Timeline.Duration())
	value, err := runtime.VM.RunString(`snapshots[snapshots.length-1].tokens[0].value === DATA.max_score && snapshots[snapshots.length-1].nodes.length === 1 && ring.opacity === 0`)
	if err != nil || !value.ToBoolean() {
		t.Fatalf("incorrect final frame: %v %v", value, err)
	}
	frame(1)
	frame(19.2)
	if first != frame(8) {
		t.Fatal("backward seek changed the rendered DP frame")
	}
	if runtime.Timeline.Duration() < 20 {
		t.Fatal("missing animation stages")
	}
}

func TestSceneSupportsOtherProblemsAndExactIntegers(t *testing.T) {
	for _, n := range []int{3, 4, 8} {
		p := polygon.Problem{Values: make([]string, n), Operators: make([]string, n)}
		for i := range p.Values {
			p.Values[i] = "9007199254740993"
			p.Operators[i] = "+"
		}
		r, err := polygon.Solve(p)
		if err != nil {
			t.Fatal(err)
		}
		source, err := Script(r)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(source, `"9007199254740993"`) {
			t.Fatal("integer precision was lost")
		}
		file := filepath.Join(t.TempDir(), "other.js")
		if err := os.WriteFile(file, []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
		runtime, err := script.LoadFile(file, 1280, 720)
		if err != nil {
			t.Fatal(err)
		}
		for _, at := range []float64{1, 8, 16, runtime.Timeline.Duration(), 2} {
			runtime.Timeline.Seek(at)
			if err := runtime.Scene.Update(at, 0); err != nil {
				t.Fatalf("n=%d t=%g: %v", n, at, err)
			}
		}
	}
	if _, err := Script(nil); err == nil {
		t.Fatal("accepted nil result")
	}
	p := polygon.Problem{Values: make([]string, 9), Operators: make([]string, 9)}
	for i := range p.Values {
		p.Values[i] = "1"
		p.Operators[i] = "+"
	}
	r, err := polygon.Solve(p)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Script(r); err == nil {
		t.Fatal("accepted too many animated vertices")
	}
}
