package script_test

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/image/font/gofont/goregular"

	"github.com/23dd/animation/mobject"
	"github.com/23dd/animation/script"
)

func TestJavaScriptBuildsSceneAndTimeline(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "scene.js")
	source := `const c=circle({radius:20,position:[10,30],fill:"#ff0000"});add(c);play(moveTo(c,[110,30]),{duration:2,ease:"smooth"});wait(.5);play(fadeOut(c));`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	r, err := script.LoadFile(path, 640, 360)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Scene.Snapshot()) != 1 {
		t.Fatalf("objects: %d", len(r.Scene.Snapshot()))
	}
	if r.Timeline.Duration() != 3.5 {
		t.Fatalf("duration: %v", r.Timeline.Duration())
	}
	r.Timeline.Seek(1)
	if x := r.Scene.Snapshot()[0].Base().Position().X; x != 60 {
		t.Fatalf("position: %v", x)
	}
}

func TestValueTrackerDrivesUpdater(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "updater.js")
	source := `const n=value(0);const c=circle();c.update((self)=>{self.x=n.value*10});add(c);play(animateValue(n,5),{duration:1});`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	r, err := script.LoadFile(path, 320, 180)
	if err != nil {
		t.Fatal(err)
	}
	r.Timeline.Seek(.5)
	r.Scene.Update(.5, .5)
	if x := r.Scene.Snapshot()[0].Base().Position().X; x != 25 {
		t.Fatalf("updater position: %v", x)
	}
}

func TestPlayAcceptsEaseWithoutDuration(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ease.js")
	source := `const c=circle();add(c);play(moveTo(c,[100,0]),{ease:"smooth"});`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	r, err := script.LoadFile(path, 320, 180)
	if err != nil {
		t.Fatal(err)
	}
	if r.Timeline.Duration() != 1 {
		t.Fatalf("duration: %v", r.Timeline.Duration())
	}
}

func TestTextResolvesCustomFontRelativeToScene(t *testing.T) {
	dir := t.TempDir()
	fontPath := filepath.Join(dir, "custom.ttf")
	if err := os.WriteFile(fontPath, goregular.TTF, 0600); err != nil {
		t.Fatal(err)
	}
	scenePath := filepath.Join(dir, "font.js")
	source := `const label=text("Custom font",{font:"./custom.ttf",size:32});add(label);`
	if err := os.WriteFile(scenePath, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	r, err := script.LoadFile(scenePath, 320, 180)
	if err != nil {
		t.Fatal(err)
	}
	label, ok := r.Scene.Snapshot()[0].(*mobject.Text)
	if !ok {
		t.Fatalf("object type: %T", r.Scene.Snapshot()[0])
	}
	if label.Font != fontPath {
		t.Fatalf("font path = %q, want %q", label.Font, fontPath)
	}
	if label.FontVersion == 0 {
		t.Fatal("font version was not recorded")
	}
}

func TestTextRejectsMissingCustomFont(t *testing.T) {
	dir := t.TempDir()
	scenePath := filepath.Join(dir, "missing-font.js")
	if err := os.WriteFile(scenePath, []byte(`add(text("Missing",{font:"./missing.ttf"}));`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := script.LoadFile(scenePath, 320, 180); err == nil {
		t.Fatal("expected missing font error")
	}
}

func TestTextRejectsInvalidCustomFont(t *testing.T) {
	dir := t.TempDir()
	fontPath := filepath.Join(dir, "invalid.ttf")
	if err := os.WriteFile(fontPath, []byte("not a font"), 0600); err != nil {
		t.Fatal(err)
	}
	scenePath := filepath.Join(dir, "invalid-font.js")
	if err := os.WriteFile(scenePath, []byte(`add(text("Invalid",{font:"./invalid.ttf"}));`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := script.LoadFile(scenePath, 320, 180); err == nil {
		t.Fatal("expected invalid font error")
	}
}

func TestEffectsAnimationsRegisterAndAnimate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "effects.js")
	source := `
const c = circle({radius: 20});
const g = group(circle(), circle(), circle());
add(c, g);
play(growFromCenter(c));
play(shrinkToCenter(c));
play(growFromPoint(c, [100, 0]));
play(growFromEdge(c, "left", 80));
play(spinIn(c));
play(spinOut(c));
play(flash(c, "#ffff00"));
play(indicate(c, 1.5, "#00ffff"));
play(blink(c, 3));
play(wiggle(c, 0.4, 3));
play(wave(g, 25));
play(zoomIn(c, 3));
play(zoomOut(c, 3));
play(focusOn(c, 1.4));
`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	r, err := script.LoadFile(path, 640, 360)
	if err != nil {
		t.Fatal(err)
	}
	// Every effect defaults to a one-second duration.
	if got := r.Timeline.Duration(); got != 14 {
		t.Fatalf("duration: %v", got)
	}
	// At t=0.5 the first effect (growFromCenter) is halfway done.
	r.Timeline.Seek(0.5)
	c := r.Scene.Snapshot()[0]
	if s := c.Base().Transform.Scale; s.X != 0.5 || s.Y != 0.5 {
		t.Fatalf("growFromCenter scale at t=0.5: %+v", s)
	}
	// Seek to the end and confirm every effect finished without error.
	r.Timeline.Seek(r.Timeline.Duration())
}

func TestDrawLineEffects(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "draw.js")
	source := `
const axis = line({start: [-100, 0], end: [100, 0]});
const nib = circle({radius: 6});
const center = line({start: [-100, 0], end: [100, 0]});
add(axis, nib, center);
play(drawLine(axis, nib), {duration: 1});
play(drawLineFromCenter(center), {duration: 1});
`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	r, err := script.LoadFile(path, 640, 360)
	if err != nil {
		t.Fatal(err)
	}
	if got := r.Timeline.Duration(); got != 2 {
		t.Fatalf("duration: %v", got)
	}
	// At t=0.5 the axis is half drawn and the nib follows the midpoint.
	r.Timeline.Seek(0.5)
	axis := r.Scene.Snapshot()[0].(*mobject.Line)
	if axis.Base().DrawProgress != 0.5 {
		t.Fatalf("draw progress: %v", axis.Base().DrawProgress)
	}
	nib := r.Scene.Snapshot()[1]
	if nib.Base().Transform.Position.X != 0 {
		t.Fatalf("nib x: %v", nib.Base().Transform.Position.X)
	}
	// drawLineFromCenter at t=1.5 grows outward from the midpoint.
	r.Timeline.Seek(1.5)
	center := r.Scene.Snapshot()[2].(*mobject.Line)
	if center.Start.X != -50 || center.End.X != 50 {
		t.Fatalf("center line endpoints: %v, %v", center.Start.X, center.End.X)
	}
}

