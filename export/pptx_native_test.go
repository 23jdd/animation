package export

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"image/color"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/23dd/animation/animation"
	"github.com/23dd/animation/geometry"
	"github.com/23dd/animation/mobject"
	"github.com/23dd/animation/scene"
	"github.com/23dd/animation/timeline"
)

func nativeTestScene() (*scene.Scene, *timeline.Timeline, *mobject.Rectangle) {
	s := scene.New(1280, 720)
	target := mobject.NewRectangle(100, 60)
	target.SetPosition(geometry.V2(-200, 0))
	target.Opacity = 0
	s.Add(target)
	tl := timeline.New()
	tl.Add(.5, animation.FadeIn(target))
	tl.Add(1.5, animation.WithTiming(animation.Parallel(
		animation.MoveTo(target, geometry.V2(200, 80)),
		animation.ScaleTo(target, geometry.V2(2, 1.5)),
		animation.RotateBy(target, math.Pi*2),
	), 2, animation.Smooth))
	tl.Add(4, animation.FadeOut(target))
	return s, tl, target
}

func TestPPTXNativeEditableAnimation(t *testing.T) {
	s, tl, target := nativeTestScene()
	tl.Seek(2)
	before := *target
	dir := t.TempDir()
	file := filepath.Join(dir, "动画 & 原生.pptx")
	if err := PPTXNative(file, s, tl, 60); err != nil {
		t.Fatal(err)
	}
	if tl.Time != 2 || target.Transform != before.Transform || target.Opacity != before.Opacity {
		t.Fatal("native export did not restore the timeline and object state")
	}
	z, err := zip.OpenReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	var slide []byte
	for _, part := range z.File {
		if strings.HasPrefix(part.Name, "ppt/media/") {
			t.Fatal("native export must not rasterize objects")
		}
		stream, err := part.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(stream)
		stream.Close()
		if err != nil {
			t.Fatal(err)
		}
		if strings.HasSuffix(part.Name, ".xml") || strings.HasSuffix(part.Name, ".rels") {
			decoder := xml.NewDecoder(bytes.NewReader(data))
			for {
				if _, err := decoder.Token(); err == io.EOF {
					break
				} else if err != nil {
					t.Fatalf("invalid XML in %s: %v", part.Name, err)
				}
			}
		}
		if part.Name == "ppt/slides/slide1.xml" {
			slide = data
		}
	}
	for _, expected := range []string{`<p:sp>`, `<p:timing>`, `nodeType="mainSeq"`, `<p:animRot`, `<p:animScale>`, `ScaleX`, `ScaleY`, `filter="fade"`, `presetClass="entr"`, `presetClass="exit"`, `advClick="0" advTm="5000"`, `ppt_x`, `ppt_y`, `delay="500"`, `delay="4000"`, `to="21600000"`} {
		if !bytes.Contains(slide, []byte(expected)) {
			t.Errorf("missing native animation content %s", expected)
		}
	}
	decoder := xml.NewDecoder(bytes.NewReader(slide))
	ids := make(map[string]bool)
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if start, ok := token.(xml.StartElement); ok && start.Name.Local == "cTn" {
			for _, attr := range start.Attr {
				if attr.Name.Local == "id" {
					if ids[attr.Value] {
						t.Fatalf("duplicate timing node ID %s", attr.Value)
					}
					ids[attr.Value] = true
				}
			}
		}
	}
}

func TestPPTXNativeSamplingAndGroups(t *testing.T) {
	s := scene.New(1280, 720)
	c := mobject.NewCircle(20)
	c.SetPosition(geometry.V2(30, 0))
	g := mobject.NewGroup(c)
	g.SetPosition(geometry.V2(100, 20))
	g.Transform.Scale = geometry.V2(2, 2)
	s.Add(g)
	tl := timeline.New()
	tl.Add(0, animation.RotateBy(g, math.Pi*2))
	d, err := sampleNativeDeck(s, tl, 60)
	if err != nil {
		t.Fatal(err)
	}
	first, last := d.tracks[0].states[0], d.tracks[0].states[len(d.times)-1]
	if first.x != 800 || first.y != 380 || first.w != 80 || math.Abs(last.rotation-360) > 1e-7 {
		t.Fatalf("group transform or full-turn rotation lost: %+v %+v", first, last)
	}
	times, err := nativeSampleTimes(tl, 7)
	if err != nil || times[0] != 0 || times[len(times)-1] != 1000 {
		t.Fatalf("sampling lost boundaries: %v %v", times, err)
	}
	for i := 1; i < len(times); i++ {
		if times[i] <= times[i-1] {
			t.Fatal("duplicate or unordered samples")
		}
	}
	for _, fps := range []float64{0, -1, 121, math.NaN(), math.Inf(1)} {
		if _, err := nativeSampleTimes(tl, fps); err == nil {
			t.Errorf("accepted fps %g", fps)
		}
	}
}

func TestPPTXNativeRejectsUnsupportedAndPreservesOutput(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*scene.Scene, *timeline.Timeline, *mobject.Rectangle)
	}{
		{"create", func(s *scene.Scene, tl *timeline.Timeline, obj *mobject.Rectangle) { tl.Add(0, animation.Create(obj)) }},
		{"fill", func(s *scene.Scene, tl *timeline.Timeline, obj *mobject.Rectangle) {
			tl.Add(0, animation.ColorTo(obj, color.RGBA{R: 240, A: 255}))
		}},
		{"size", func(s *scene.Scene, tl *timeline.Timeline, obj *mobject.Rectangle) {
			obj.AddUpdater(func(_ mobject.Mobject, at, _ float64) error { obj.Width = 100 + at; return nil })
		}},
		{"partial opacity", func(s *scene.Scene, tl *timeline.Timeline, obj *mobject.Rectangle) {
			tl.Add(0, animation.OpacityTo(obj, .5))
		}},
		{"negative scale", func(s *scene.Scene, tl *timeline.Timeline, obj *mobject.Rectangle) {
			obj.Transform.Scale = geometry.V2(-1, -1)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := scene.New(1280, 720)
			obj := mobject.NewRectangle(100, 60)
			s.Add(obj)
			tl := timeline.New()
			tl.Wait(2)
			tc.change(s, tl, obj)
			tl.Seek(1.75)
			file := filepath.Join(t.TempDir(), "existing.pptx")
			if err := os.WriteFile(file, []byte("original"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := PPTXNative(file, s, tl, 30); err == nil {
				t.Fatal("accepted unsupported animation")
			}
			data, _ := os.ReadFile(file)
			if string(data) != "original" || tl.Time != 1.75 {
				t.Fatal("failed export changed original file or timeline")
			}
		})
	}
}

func TestNativeFadesAndKeyframes(t *testing.T) {
	states := []nativeState{{alpha: 0}, {alpha: 0}, {alpha: .2}, {alpha: .5}, {alpha: .5}, {alpha: .3}, {alpha: 0}}
	times := []int{0, 500, 750, 1000, 1500, 1750, 2000}
	peak, fades, err := nativeFades(states, times)
	if err != nil || peak != .5 || len(fades) != 2 || fades[0] != (nativeFade{500, 1000, true}) || fades[1] != (nativeFade{1500, 2000, false}) {
		t.Fatalf("incorrect fade scheduling: %g %+v %v", peak, fades, err)
	}
	linear := nativeKeyIndices([]float64{0, 1, 2, 3}, []int{0, 100, 200, 300}, .0001)
	if len(linear) != 2 || linear[0] != 0 || linear[1] != 3 {
		t.Fatal("linear samples were not reduced")
	}
	plateau := nativeKeyIndices([]float64{0, 0, 1, 1}, []int{0, 100, 200, 300}, .0001)
	if len(plateau) != 4 {
		t.Fatal("wait intervals were lost")
	}
}

func TestNativeExitFadeLatchesHiddenWithoutExtendingDuration(t *testing.T) {
	s, tl := scene.New(1280, 720), timeline.New()
	c := mobject.NewCircle(30)
	s.Add(c)
	tl.Add(0, animation.WithTiming(animation.FadeOut(c), .8, animation.Linear))
	d, err := sampleNativeDeck(s, tl, 60)
	if err != nil {
		t.Fatal(err)
	}
	slide, err := d.slideXML()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(slide, `delay="799"`) || !strings.Contains(slide, `val="hidden"`) {
		t.Fatal("exit effect must latch visibility on its final millisecond")
	}
	if strings.Contains(slide, `style.opacity`) {
		t.Fatal("a fade filter must not inherit a numeric property animation")
	}
}

func TestNativeTextAndPaths(t *testing.T) {
	s := scene.New(1280, 720)
	label := mobject.NewText("中文 < & >\nEditable", 36)
	var p geometry.Path
	p.MoveTo(geometry.V2(-40, 0)).CubicTo(geometry.V2(-20, -80), geometry.V2(20, 80), geometry.V2(40, 0))
	s.Add(label, mobject.NewPath(p), mobject.NewLine(geometry.V2(-10, 20), geometry.V2(10, 20)))
	d, err := sampleNativeDeck(s, timeline.New(), 60)
	if err != nil {
		t.Fatal(err)
	}
	slide, err := d.slideXML()
	if err != nil || !strings.Contains(slide, "中文 &lt; &amp; &gt;") || !strings.Contains(slide, `<a:cubicBezTo>`) || !strings.Contains(slide, "<p:txBody>") {
		t.Fatalf("editable text/path missing: %v", err)
	}
	if strings.Contains(slide, `advTm="0"`) {
		t.Fatal("zero-duration slide must remain visible until advanced")
	}
}
