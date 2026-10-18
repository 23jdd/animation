package app

import (
	"archive/zip"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestPPTXSelectionFlags(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want []float64
	}{
		{nil, []float64{0}},
		{[]string{"-time", "1.5"}, []float64{1.5}},
		{[]string{"-times", "2, 0, 1, 1"}, []float64{2, 0, 1, 1}},
		{[]string{"-interval", ".8"}, []float64{0, .8, 1.6, 2}},
		{[]string{"-interval", "5"}, []float64{0, 2}},
	} {
		o, _, err := flags("pptx", tc.args)
		if err != nil {
			t.Fatal(err)
		}
		if o.scale != 2 {
			t.Fatal("PPTX should default to 2x raster output")
		}
		got, err := pptxFrameTimes(o, 2)
		if err != nil || !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("args=%v got=%v want=%v err=%v", tc.args, got, tc.want, err)
		}
	}
	for _, args := range [][]string{{"-time", "0", "-times", "0"}, {"-times", "1", "-interval", "1"}, {"-time", "0", "-interval", "1"}} {
		if _, _, err := flags("pptx", args); err == nil {
			t.Errorf("accepted conflicting PPTX flags %v", args)
		}
	}
}

func TestPPTXCommandsExportBitmapFrames(t *testing.T) {
	dir := t.TempDir()
	asset := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for i := 0; i < len(asset.Pix); i += 4 {
		asset.Pix[i], asset.Pix[i+3] = 255, 255
	}
	f, err := os.Create(filepath.Join(dir, "asset.png"))
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, asset); err != nil {
		f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(dir, "scene.js")
	if err := os.WriteFile(source, []byte(`add(image("asset.png",{width:20}));wait(1);`), 0600); err != nil {
		t.Fatal(err)
	}
	for _, command := range []string{"ppt", "pptx"} {
		output := filepath.Join(dir, command+".pptx")
		if err := Run([]string{command, "-width", "80", "-height", "40", "-times", "0,.5", source, output}, io.Discard); err != nil {
			t.Fatal(err)
		}
		z, err := zip.OpenReader(output)
		if err != nil {
			t.Fatal(err)
		}
		images := 0
		for _, part := range z.File {
			if !strings.HasPrefix(part.Name, "ppt/media/") {
				continue
			}
			images++
			stream, err := part.Open()
			if err != nil {
				t.Fatal(err)
			}
			frame, err := png.Decode(stream)
			stream.Close()
			if err != nil {
				t.Fatal(err)
			}
			if frame.Bounds() != image.Rect(0, 0, 160, 80) {
				t.Fatalf("PPTX default 2x output not applied: %v", frame.Bounds())
			}
			if got := color.RGBAModel.Convert(frame.At(80, 40)).(color.RGBA); got.R < 250 || got.G > 5 {
				t.Fatalf("bitmap image was lost in PPTX export: %v", got)
			}
		}
		z.Close()
		if images != 2 {
			t.Fatalf("expected two frame images, got %d", images)
		}
	}
}

func TestPPTXRejectsInvalidSelections(t *testing.T) {
	for _, args := range [][]string{
		{"-times", ""}, {"-times", "1,"}, {"-times", "NaN"}, {"-times", "-1"}, {"-times", "3"},
		{"-time", "NaN"}, {"-time", "-1"}, {"-time", "3"},
		{"-interval", "0"}, {"-interval", "-1"}, {"-interval", "NaN"}, {"-interval", "Inf"}, {"-interval", ".0001"},
	} {
		o, _, err := flags("pptx", args)
		if err == nil {
			_, err = pptxFrameTimes(o, 2)
		}
		if err == nil {
			t.Errorf("accepted invalid PPTX selection %v", args)
		}
	}
	for _, p := range []struct{ duration, interval float64 }{{0, 1}, {.3, .1}, {5e-10, 2e-10}, {499, 1}} {
		got, err := pptxFrameTimes(options{pptxSelection: "interval", pptxInterval: p.interval}, p.duration)
		if err != nil || len(got) > 500 || got[0] != 0 || math.Abs(got[len(got)-1]-p.duration) > 1e-20 {
			t.Fatalf("sampling duration=%g interval=%g: %v %v", p.duration, p.interval, got, err)
		}
	}
}

func TestPPTXCommandAliasesAndExtension(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "scene.js")
	if err := os.WriteFile(path, []byte(`add(circle());wait(1);`), 0600); err != nil {
		t.Fatal(err)
	}
	for _, command := range []string{"ppt", "pptx"} {
		err := Run([]string{command, path, filepath.Join(dir, "unsupported.ppt")}, io.Discard)
		if err == nil || !strings.Contains(err.Error(), ".pptx extension") {
			t.Fatalf("command %s not routed to PPTX: %v", command, err)
		}
	}
}
