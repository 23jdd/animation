package export

import (
	"fmt"
	"github.com/23dd/animation/renderer"
	"github.com/23dd/animation/scene"
	"github.com/23dd/animation/timeline"
	"image/png"
	"os"
)

func PNG(path string, s *scene.Scene, t *timeline.Timeline, r renderer.Renderer, at float64) error {
	t.Seek(at)
	s.Update(t.Time, 0)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if err = png.Encode(f, r.Render(s)); err != nil {
		return fmt.Errorf("encode png: %w", err)
	}
	return f.Close()
}
