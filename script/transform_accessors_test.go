package script_test

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/23dd/animation/script"
)

func TestUpdaterCanScaleAndMirrorGroupWhenSeeking(t *testing.T) {
	path := filepath.Join(t.TempDir(), "scale-loop.js")
	source := `
const rig = group(rect({width: 20, height: 10}));
rig.scaleX = 0.5;
rig.scaleY = -0.75;
if (rig.scaleX !== 0.5 || rig.scaleY !== -0.75) throw new Error("scale accessors");
rig.update((self, time) => {
  const theta = (time % 10) * Math.PI / 5;
  self.scaleX = 1 + 0.1 * Math.sin(theta);
  self.scaleY = -(0.6 + 0.05 * Math.cos(theta));
});
add(rig);
wait(10);`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	runtime, err := script.LoadFile(path, 320, 180)
	if err != nil {
		t.Fatal(err)
	}
	for _, at := range []float64{0, 2.5, 7.5, 10, 2.5, 0} {
		runtime.Timeline.Seek(at)
		if err := runtime.Scene.Update(at, 0); err != nil {
			t.Fatal(err)
		}
		theta := math.Mod(at, 10) * math.Pi / 5
		got := runtime.Scene.Snapshot()[0].Base().Transform.Scale
		if math.Abs(got.X-(1+0.1*math.Sin(theta))) > 1e-12 ||
			math.Abs(got.Y+(0.6+0.05*math.Cos(theta))) > 1e-12 {
			t.Fatalf("at %g: scale = %+v", at, got)
		}
	}
}
