package app

import (
	"errors"
	"flag"
	"fmt"
	"io"
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
	_, err := fmt.Fprintln(w, "Animation engine\n\nUsage:\n  anim preview [options] scene.js\n  anim render  [options] scene.js output.png\n  anim gif     [options] scene.js output.gif\n  anim mp4     [options] scene.js output.mp4\n\nOptions: -width 1280 -height 720 -fps 60 -time <seconds>")
	return err
}

type options struct {
	width, height int
	fps, at       float64
}

func flags(name string, args []string) (options, []string, error) {
	var o options
	f := flag.NewFlagSet(name, flag.ContinueOnError)
	f.SetOutput(io.Discard)
	f.IntVar(&o.width, "width", 1280, "frame width")
	f.IntVar(&o.height, "height", 720, "frame height")
	f.Float64Var(&o.fps, "fps", 60, "frames per second")
	f.Float64Var(&o.at, "time", 0, "render time")
	if err := f.Parse(args); err != nil {
		return o, nil, err
	}
	if o.width <= 0 || o.height <= 0 {
		return o, nil, errors.New("width and height must be positive")
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
	renderer := ggrenderer.New(o.width, o.height)
	switch kind {
	case "png":
		return animexport.PNG(rest[1], rt.Scene, rt.Timeline, renderer, o.at)
	case "gif":
		return animexport.GIF(rest[1], rt.Scene, rt.Timeline, renderer, o.fps)
	default:
		return animexport.MP4(rest[1], rt.Scene, rt.Timeline, renderer, o.fps)
	}
}
