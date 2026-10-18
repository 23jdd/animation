package script

import (
	"fmt"
	"github.com/23dd/animation/scene"
	"github.com/23dd/animation/timeline"
	"github.com/dop251/goja"
	"image"
	"os"
	"path/filepath"
)

type Runtime struct {
	VM       *goja.Runtime
	Scene    *scene.Scene
	Timeline *timeline.Timeline
	BaseDir  string
	cursor   float64
	callErr  error
	images   map[string]image.Image
}

func New(width, height int) (*Runtime, error) {
	baseDir, _ := os.Getwd()
	r := &Runtime{VM: goja.New(), Scene: scene.New(width, height), Timeline: timeline.New(), BaseDir: baseDir}
	if err := r.register(); err != nil {
		return nil, err
	}
	return r, nil
}
func LoadFile(path string, width, height int) (*Runtime, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	code, err := Load(abs)
	if err != nil {
		return nil, err
	}
	r, err := New(width, height)
	if err != nil {
		return nil, err
	}
	r.BaseDir = filepath.Dir(abs)
	if _, err = r.VM.RunScript(abs, code); err != nil {
		return nil, fmt.Errorf("script %s: %w", abs, err)
	}
	r.Timeline.Seek(0)
	return r, nil
}
