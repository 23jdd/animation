package export

import (
	"bytes"
	"errors"
	"fmt"
	"image/png"
	"io"
	"math"
	"path/filepath"
	"strings"

	"github.com/23dd/animation/renderer"
	"github.com/23dd/animation/scene"
	"github.com/23dd/animation/timeline"
)

const MaxPPTXSlides = 500

// PPTX exports static animation frames as full-slide PNG pictures. An empty
// times slice selects time zero. Order and repeated times are preserved.
// The result is PPTX, not legacy binary PPT or native PowerPoint animation.
func PPTX(path string, s *scene.Scene, t *timeline.Timeline, r renderer.Renderer, times []float64) error {
	return PPTXWithProgress(path, s, t, r, times, nil)
}

func PPTXWithProgress(path string, s *scene.Scene, t *timeline.Timeline, r renderer.Renderer, times []float64, progress func(done, total int)) error {
	if strings.ToLower(filepath.Ext(path)) != ".pptx" {
		return errors.New("PPTX output must have a .pptx extension (legacy .ppt is not supported)")
	}
	if s == nil || t == nil || r == nil {
		return errors.New("PPTX export requires a scene, timeline and renderer")
	}
	times, err := validatePPTXTimes(times, t.Duration())
	if err != nil {
		return err
	}
	w, h := r.Size()
	sw, sh, err := pptxSlideSize(w, h)
	if err != nil {
		return err
	}
	title := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	return savePPTX(path, func(out io.Writer) error {
		archive, err := newPPTXArchive(out, sw, sh, title, len(times))
		if err != nil {
			return err
		}
		err = renderPPTXFrames(s, t, r, times, archive.addFrame, progress)
		return errors.Join(err, archive.zip.Close())
	})
}

func validatePPTXTimes(times []float64, duration float64) ([]float64, error) {
	if duration < 0 || math.IsNaN(duration) || math.IsInf(duration, 0) {
		return nil, errors.New("PPTX timeline duration must be non-negative and finite")
	}
	if len(times) == 0 {
		times = []float64{0}
	}
	if len(times) > MaxPPTXSlides {
		return nil, fmt.Errorf("PPTX supports at most %d slides per export", MaxPPTXSlides)
	}
	for i, at := range times {
		if at < 0 || at > duration || math.IsNaN(at) || math.IsInf(at, 0) {
			return nil, fmt.Errorf("PPTX slide %d: time must be finite and between 0 and %g seconds", i+1, duration)
		}
	}
	return append([]float64(nil), times...), nil
}

func pptxSlideSize(w, h int) (int32, int32, error) {
	if w <= 0 || h <= 0 {
		return 0, 0, errors.New("PPTX canvas dimensions must be positive")
	}
	// Normal canvas: 13 1/3 inches on the long side. Keep the exact canvas
	// ratio in portrait and landscape, within OOXML's 1–56 inch size limits.
	long, short := math.Max(float64(w), float64(h)), math.Min(float64(w), float64(h))
	longEMU := math.Max(12192000, 914400*long/short)
	if longEMU > 51206400 {
		return 0, 0, errors.New("PPTX canvas aspect ratio exceeds PowerPoint's slide size limits")
	}
	shortEMU := longEMU * short / long
	if w >= h {
		return int32(math.Round(longEMU)), int32(math.Round(shortEMU)), nil
	}
	return int32(math.Round(shortEMU)), int32(math.Round(longEMU)), nil
}

// Every PNG has an independent buffer. The package writer streams it directly
// into the ZIP; it never retains all slide images or writes intermediate PNGs.
func renderPPTXFrames(s *scene.Scene, t *timeline.Timeline, r renderer.Renderer, times []float64, add func(int, float64, []byte) error, progress func(int, int)) (result error) {
	previous := t.Time
	defer func() {
		t.Seek(previous)
		if err := s.Update(t.Time, 0); err != nil {
			result = errors.Join(result, fmt.Errorf("restore scene after PPTX export: %w", err))
		}
	}()
	w, h := r.Size()
	for i, at := range times {
		t.Seek(at)
		if err := s.Update(t.Time, 0); err != nil {
			return fmt.Errorf("update PPTX slide %d at %gs: %w", i+1, at, err)
		}
		frame := r.Render(s)
		if frame == nil || frame.Bounds().Empty() {
			return fmt.Errorf("PPTX slide %d: renderer returned an empty image", i+1)
		}
		bounds := frame.Bounds()
		if float64(bounds.Dx())*float64(bounds.Dy()) > 64_000_000 {
			return errors.New("PPTX frame exceeds 64 million pixels; reduce -scale")
		}
		// Integer output sizes can differ by a pixel after scaling/rounding.
		if math.Abs(float64(bounds.Dy())-float64(bounds.Dx())*float64(h)/float64(w)) > 1 {
			return errors.New("PPTX renderer output must preserve the logical canvas aspect ratio")
		}
		var buf bytes.Buffer
		if err := png.Encode(&buf, frame); err != nil {
			return fmt.Errorf("encode PPTX slide %d: %w", i+1, err)
		}
		if err := add(i, at, buf.Bytes()); err != nil {
			return fmt.Errorf("PPTX slide %d at %gs: %w", i+1, at, err)
		}
		if progress != nil {
			progress(i+1, len(times))
		}
	}
	return nil
}
