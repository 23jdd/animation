package app

import (
	"image/gif"
	"os"
	"path/filepath"
	"testing"
)

func TestGIFScalePreservesSceneLayout(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "scene.js")
	if err := os.WriteFile(script, []byte(`add(rect({width: 20, height: 10, fill: "#ffffff"}));`), 0600); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(dir, "scaled.gif")
	if err := runRender([]string{"-width", "80", "-height", "40", "-scale", "2", script, output}, "gif"); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(output)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	img, err := gif.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	if img.Bounds().Dx() != 160 || img.Bounds().Dy() != 80 {
		t.Fatalf("unexpected scaled bounds: %v", img.Bounds())
	}
	// The 20-unit rectangle must grow to 40 pixels, centered in the output.
	for _, x := range []int{65, 95} {
		r, g, b, _ := img.At(x, 40).RGBA()
		if r < 60000 || g < 60000 || b < 60000 {
			t.Fatalf("scaled rectangle missing at x=%d", x)
		}
	}
}

func TestInvalidExportScale(t *testing.T) {
	for _, scale := range []string{"0", "-1", "NaN", "+Inf"} {
		if _, _, err := flags("gif", []string{"-scale", scale}); err == nil {
			t.Errorf("accepted invalid scale %s", scale)
		}
	}
}
