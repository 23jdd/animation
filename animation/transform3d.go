package animation

import (
	"github.com/23dd/animation/geometry"
	"github.com/23dd/animation/mobject"
	"github.com/23dd/animation/scene"
)

func Move3DTo(target *mobject.Mesh3D, to geometry.Vec3) Animation {
	var from geometry.Vec3
	return New(1, Linear,
		func(float64) { from = target.Transform3D.Position },
		func(t float64) { target.Transform3D.Position = from.Lerp(to, t) }, nil,
		func(float64) { target.Transform3D.Position = from })
}

func Move3DBy(target *mobject.Mesh3D, delta geometry.Vec3) Animation {
	var from geometry.Vec3
	return New(1, Linear,
		func(float64) { from = target.Transform3D.Position },
		func(t float64) { target.Transform3D.Position = from.Add(delta.Mul(t)) }, nil,
		func(float64) { target.Transform3D.Position = from })
}

func Rotate3DTo(target *mobject.Mesh3D, to geometry.Vec3) Animation {
	var from geometry.Vec3
	return New(1, Linear,
		func(float64) { from = target.Transform3D.Rotation },
		func(t float64) { target.Transform3D.Rotation = from.Lerp(to, t) }, nil,
		func(float64) { target.Transform3D.Rotation = from })
}

func Rotate3DBy(target *mobject.Mesh3D, delta geometry.Vec3) Animation {
	var from geometry.Vec3
	return New(1, Linear,
		func(float64) { from = target.Transform3D.Rotation },
		func(t float64) { target.Transform3D.Rotation = from.Add(delta.Mul(t)) }, nil,
		func(float64) { target.Transform3D.Rotation = from })
}

func Scale3DTo(target *mobject.Mesh3D, to geometry.Vec3) Animation {
	var from geometry.Vec3
	return New(1, Linear,
		func(float64) { from = target.Transform3D.Scale },
		func(t float64) { target.Transform3D.Scale = from.Lerp(to, t) }, nil,
		func(float64) { target.Transform3D.Scale = from })
}

func Scale3DBy(target *mobject.Mesh3D, factor geometry.Vec3) Animation {
	var from geometry.Vec3
	return New(1, Linear,
		func(float64) { from = target.Transform3D.Scale },
		func(t float64) {
			target.Transform3D.Scale = geometry.V3(
				from.X*(1+(factor.X-1)*t),
				from.Y*(1+(factor.Y-1)*t),
				from.Z*(1+(factor.Z-1)*t),
			)
		}, nil,
		func(float64) { target.Transform3D.Scale = from })
}

func MoveCamera3DTo(camera *scene.Camera3D, to geometry.Vec3) Animation {
	var from geometry.Vec3
	return New(1, Linear,
		func(float64) { from = camera.Position },
		func(t float64) { camera.Position = from.Lerp(to, t) }, nil,
		func(float64) { camera.Position = from })
}

func LookAt3D(camera *scene.Camera3D, to geometry.Vec3) Animation {
	var from geometry.Vec3
	return New(1, Linear,
		func(float64) { from = camera.Target },
		func(t float64) { camera.Target = from.Lerp(to, t) }, nil,
		func(float64) { camera.Target = from })
}

func CameraFOVTo(camera *scene.Camera3D, to float64) Animation {
	var from float64
	return New(1, Linear,
		func(float64) { from = camera.FOV },
		func(t float64) { camera.FOV = from + (to-from)*t }, nil,
		func(float64) { camera.FOV = from })
}
