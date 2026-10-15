package renderer

import (
	"github.com/23dd/animation/scene"
	"image"
)

type Renderer interface {
	Render(*scene.Scene) image.Image
	Size() (int, int)
}

// OutputSizer is implemented by renderers that can rasterize the logical
// canvas directly at a different output resolution.
type OutputSizer interface {
	SetOutputSize(width, height int)
}
