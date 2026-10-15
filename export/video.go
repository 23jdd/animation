package export

import (
	"fmt"
	"image"
	"image/draw"
	"os/exec"
	"strconv"

	"github.com/23dd/animation/renderer"
	"github.com/23dd/animation/scene"
	"github.com/23dd/animation/timeline"
)

func MP4(path string, s *scene.Scene, t *timeline.Timeline, r renderer.Renderer, fps float64) error {
	if fps <= 0 {
		fps = 60
	}
	w, h := r.Size()
	cmd := exec.Command("ffmpeg", "-y", "-f", "rawvideo", "-pixel_format", "rgba", "-video_size", fmt.Sprintf("%dx%d", w, h), "-framerate", strconv.FormatFloat(fps, 'f', -1, 64), "-i", "-", "-an", "-c:v", "libx264", "-pix_fmt", "yuv420p", path)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	var stderr []byte
	cmd.Stderr = &byteWriter{p: &stderr}
	if err = cmd.Start(); err != nil {
		return fmt.Errorf("start ffmpeg: %w", err)
	}
	frames := int(t.Duration()*fps) + 1
	for i := 0; i < frames; i++ {
		t.Seek(float64(i) / fps)
		s.Update(t.Time, 1/fps)
		src := r.Render(s)
		dst := image.NewRGBA(image.Rect(0, 0, w, h))
		draw.Draw(dst, dst.Bounds(), src, src.Bounds().Min, draw.Src)
		if _, err = stdin.Write(dst.Pix); err != nil {
			_ = stdin.Close()
			_ = cmd.Wait()
			return err
		}
	}
	_ = stdin.Close()
	if err = cmd.Wait(); err != nil {
		return fmt.Errorf("ffmpeg: %w: %s", err, string(stderr))
	}
	return nil
}

type byteWriter struct{ p *[]byte }

func (w *byteWriter) Write(p []byte) (int, error) { *w.p = append(*w.p, p...); return len(p), nil }
