package scene

import (
	"github.com/23dd/animation/geometry"
	"image/color"
)

type Camera struct {
	Width, Height int
	Position      geometry.Vec2
	Zoom          float64
	Background    color.RGBA
}

func NewCamera(width, height int) *Camera {
	return &Camera{Width: width, Height: height, Zoom: 1, Background: color.RGBA{R: 18, G: 18, B: 24, A: 255}}
}
