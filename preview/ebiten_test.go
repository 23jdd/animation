package preview

import (
	"math"
	"testing"
)

func TestFormatTime(t *testing.T) {
	tests := map[float64]string{0: "00:00.000", 1.25: "00:01.250", 61.002: "01:01.002"}
	for input, want := range tests {
		if got := formatTime(input); got != want {
			t.Fatalf("formatTime(%v) = %q, want %q", input, got, want)
		}
	}
}

func TestFitCanvasPreservesAspectRatio(t *testing.T) {
	x, y, scale := fitCanvas(1280, 720, 1000, 500)
	if math.Abs(scale-500.0/720.0) > 0.0001 {
		t.Fatalf("scale = %v", scale)
	}
	if x <= 0 || y != 0 {
		t.Fatalf("offset = (%v, %v)", x, y)
	}
	x, y, scale = fitCanvas(1280, 720, 1600, 1000)
	if math.Abs(scale-1.25) > 0.0001 || x != 0 || y != 50 {
		t.Fatalf("fit = (%v, %v, %v)", x, y, scale)
	}
}

func TestFitCanvasPixelsUsesAvailableResolution(t *testing.T) {
	x, y, width, height := fitCanvasPixels(1280, 720, 1000, 500)
	if x != 55 || y != 0 || width != 889 || height != 500 {
		t.Fatalf("fit pixels = (%d, %d, %d, %d)", x, y, width, height)
	}
	x, y, width, height = fitCanvasPixels(1280, 720, 1600, 1000)
	if x != 0 || y != 50 || width != 1600 || height != 900 {
		t.Fatalf("fit pixels = (%d, %d, %d, %d)", x, y, width, height)
	}
}
