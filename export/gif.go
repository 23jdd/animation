package export

import (
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"os"

	"github.com/23dd/animation/renderer"
	"github.com/23dd/animation/scene"
	"github.com/23dd/animation/timeline"
)

func GIF(path string, s *scene.Scene, t *timeline.Timeline, r renderer.Renderer, fps float64) error {
	if fps <= 0 {
		fps = 30
	}
	duration := t.Duration()
	frames := int(duration*fps) + 1
	out := &gif.GIF{LoopCount: 0}
	for i := 0; i < frames; i++ {
		t.Seek(float64(i) / fps)
		s.Update(t.Time, 1/fps)
		src := r.Render(s)
		dst := image.NewPaletted(src.Bounds(), palette.Plan9)
		draw.FloydSteinberg.Draw(dst, dst.Rect, src, src.Bounds().Min)
		out.Image = append(out.Image, dst)
		out.Delay = append(out.Delay, int(100/fps))
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return gif.EncodeAll(f, out)
}
