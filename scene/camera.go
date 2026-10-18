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
	ThreeD        Camera3D
}

func NewCamera(width, height int) *Camera {
	return &Camera{
		Width: width, Height: height, Zoom: 1,
		Background: color.RGBA{R: 18, G: 18, B: 24, A: 255},
		ThreeD:     NewCamera3D(),
	}
}

type Camera3D struct {
	Position       geometry.Vec3
	Target         geometry.Vec3
	Up             geometry.Vec3
	FOV            float64
	Near           float64
	LightDirection geometry.Vec3
	Ambient        float64
	Diffuse        float64
	FogColor       color.RGBA
	FogNear        float64
	FogFar         float64
	FogStrength    float64
}

func NewCamera3D() Camera3D {
	return Camera3D{
		Position:       geometry.V3(0, 0, 800),
		Target:         geometry.V3(0, 0, 0),
		Up:             geometry.V3(0, 1, 0),
		FOV:            50,
		Near:           1,
		LightDirection: geometry.V3(-0.45, 0.7, 1).Normalize(),
		Ambient:        0.18,
		Diffuse:        0.82,
		FogColor:       color.RGBA{R: 18, G: 18, B: 24, A: 255},
		FogNear:        700,
		FogFar:         1800,
	}
}
