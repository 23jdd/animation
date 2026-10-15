package app

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/23dd/animation/engine"
	animexport "github.com/23dd/animation/export"
	"github.com/23dd/animation/hotreload"
	"github.com/23dd/animation/preview"
	ggrenderer "github.com/23dd/animation/renderer/gg"
	"github.com/23dd/animation/script"
)

func Run(args []string, out io.Writer) error {
	if len(args) == 0 {
		return usage(out)
	}
	switch args[0] {
	case "preview":
		return runPreview(args[1:])
	case "render":
		return runRender(args[1:], "png")
	case "gif":
		return runRender(args[1:], "gif")
	case "mp4":
		return runRender(args[1:], "mp4")
	case "help", "-h", "--help":
		return usage(out)
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}
func usage(w io.Writer) error {
	_, err := fmt.Fprintln(w, "Animation engine\n\nUsage:\n  anim preview [options] scene.js\n  anim render  [options] scene.js output.png\n  anim gif     [options] scene.js output.gif\n  anim mp4     [options] scene.js output.mp4\n\nOptions: -width 1280 -height 720 -fps 60 -time <seconds>\nExport only: -scale 1 (pixel multiplier; preserves scene layout)")
	return err
}

type options struct {
	width, height int
	fps, at       float64
	scale         float64
}

func flags(name string, args []string) (options, []string, error) {
	var o options
	f := flag.NewFlagSet(name, flag.ContinueOnError)
	f.SetOutput(io.Discard)
	f.IntVar(&o.width, "width", 1280, "frame width")
	f.IntVar(&o.height, "height", 720, "frame height")
	f.Float64Var(&o.fps, "fps", 60, "frames per second")
	f.Float64Var(&o.at, "time", 0, "render time")
	o.scale = 1
	if name != "preview" {
		f.Float64Var(&o.scale, "scale", 1, "export pixel scale (preserves scene layout)")
	}
	if err := f.Parse(args); err != nil {
		return o, nil, err
	}
	if o.width <= 0 || o.height <= 0 {
		return o, nil, errors.New("width and height must be positive")
	}
	if o.scale <= 0 || math.IsNaN(o.scale) || math.IsInf(o.scale, 0) {
		return o, nil, errors.New("scale must be positive and finite")
	}
	return o, f.Args(), nil
}
func runPreview(args []string) error {
	o, rest, err := flags("preview", args)
	if err != nil {
		return err
	}
	if len(rest) != 1 {
		return errors.New("preview requires a scene.js file")
	}
	rt, err := script.LoadFile(rest[0], o.width, o.height)
	if err != nil {
		return err
	}
	renderer := ggrenderer.New(o.width, o.height)
	e := engine.New(rt.Scene, rt.Timeline, renderer)
	reloader := &hotreload.Reloader{Path: rest[0], Width: o.width, Height: o.height, Engine: e}
	watcher, err := hotreload.New(rest[0], 150*time.Millisecond, reloader.Reload)
	if err != nil {
		return err
	}
	defer watcher.Close()
	return preview.Run(e, "Animation - "+filepath.Base(rest[0]))
}
func runRender(args []string, kind string) error {
	o, rest, err := flags(kind, args)
	if err != nil {
		return err
	}
	if len(rest) != 2 {
		return fmt.Errorf("%s requires scene.js and an output file", kind)
	}
	rt, err := script.LoadFile(rest[0], o.width, o.height)
	if err != nil {
		return err
	}
	outputWidth := math.Round(float64(o.width) * o.scale)
	outputHeight := math.Round(float64(o.height) * o.scale)
	if outputWidth < 1 || outputHeight < 1 || outputWidth > 65535 || outputHeight > 65535 {
		return errors.New("scaled dimensions must be between 1 and 65535 pixels")
	}
	if kind == "mp4" && (int(outputWidth)%2 != 0 || int(outputHeight)%2 != 0) {
		return errors.New("MP4 output width and height must be even")
	}
	renderer := ggrenderer.New(o.width, o.height)
	renderer.SetOutputSize(int(outputWidth), int(outputHeight))
	switch kind {
	case "png":
		return animexport.PNG(rest[1], rt.Scene, rt.Timeline, renderer, o.at)
	case "gif":
		started, last := time.Now(), time.Time{}
		return animexport.GIFWithProgress(rest[1], rt.Scene, rt.Timeline, renderer, o.fps, func(done, total int) {
			if done == total || time.Since(last) >= time.Second {
				fmt.Fprintf(os.Stderr, "\rGIF %d/%d (%d%%)  elapsed %s", done, total, done*100/total, time.Since(started).Round(time.Second))
				last = time.Now()
				if done == total {
					fmt.Fprintln(os.Stderr)
				}
			}
		})
	default:
		return animexport.MP4(rest[1], rt.Scene, rt.Timeline, renderer, o.fps)
	}
}
