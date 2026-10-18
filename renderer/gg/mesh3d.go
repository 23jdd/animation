package gg

import (
	"image"
	"image/color"
	"math"
	"sort"

	"github.com/23dd/animation/geometry"
	"github.com/23dd/animation/mobject"
	"github.com/23dd/animation/scene"
)

type projectedFace struct {
	points      []geometry.Vec2
	depth       float64
	fill        color.RGBA
	stroke      color.RGBA
	strokeWidth float64
	alpha       float64
	wireframe   bool
	glowRadius  float64
	glowColor   color.RGBA
}

type cameraBasis struct {
	position geometry.Vec3
	forward  geometry.Vec3
	right    geometry.Vec3
	up       geometry.Vec3
	focal    float64
	near     float64
}

type rasterVertex struct {
	screen geometry.Vec2
	world  geometry.Vec3
	normal geometry.Vec3
	depth  float64
}

type rasterTriangle struct {
	vertices      [3]rasterVertex
	faceNormal    geometry.Vec3
	fill          color.RGBA
	alpha         float64
	depth         float64
	shading       bool
	smoothShading bool
	emissive      float64
	specular      float64
	shininess     float64
}

type clipVertex struct {
	world  geometry.Vec3
	normal geometry.Vec3
	depth  float64
}

func newCameraBasis(camera *scene.Camera, logicalHeight int) cameraBasis {
	c := &camera.ThreeD
	forward := c.Target.Sub(c.Position).Normalize()
	if forward.Length() == 0 {
		forward = geometry.V3(0, 0, -1)
	}
	up := c.Up.Normalize()
	if up.Length() == 0 {
		up = geometry.V3(0, 1, 0)
	}
	right := forward.Cross(up).Normalize()
	if right.Length() == 0 {
		right = geometry.V3(1, 0, 0)
	}
	up = right.Cross(forward).Normalize()
	fov := math.Max(1, math.Min(179, c.FOV)) * math.Pi / 180
	return cameraBasis{
		position: c.Position,
		forward:  forward,
		right:    right,
		up:       up,
		focal:    (float64(logicalHeight) / 2) / math.Tan(fov/2),
		near:     math.Max(0.001, c.Near),
	}
}

func (c cameraBasis) project(world geometry.Vec3) (geometry.Vec2, float64, bool) {
	rel := world.Sub(c.position)
	depth := rel.Dot(c.forward)
	if depth <= c.near {
		return geometry.Vec2{}, depth, false
	}
	return geometry.V2(rel.Dot(c.right)*c.focal/depth, -rel.Dot(c.up)*c.focal/depth), depth, true
}

func (r *Renderer) drawMeshes3D(meshes []*mobject.Mesh3D, parentAlpha float64, applyBase bool) {
	if r.camera == nil || len(meshes) == 0 {
		return
	}
	basis := newCameraBasis(r.camera, r.height)
	triangles := make([]rasterTriangle, 0)
	overlays := make([]projectedFace, 0)
	for _, mesh := range meshes {
		triangles = append(triangles, r.rasterTriangles(mesh, basis, parentAlpha, applyBase)...)
		b := mesh.Base()
		if mesh.Wireframe || mesh.GlowRadius > 0 || (b.Stroke.A > 0 && b.StrokeWidth > 0) {
			faces := r.projectMesh(mesh, basis, parentAlpha, applyBase)
			for i := range faces {
				faces[i].wireframe = true
			}
			overlays = append(overlays, faces...)
		}
	}
	if target, ok := r.dc.Image().(*image.RGBA); ok && len(triangles) > 0 {
		r.rasterizeTriangles(target, triangles)
	}
	sort.SliceStable(overlays, func(i, j int) bool { return overlays[i].depth > overlays[j].depth })
	for _, face := range overlays {
		r.drawProjectedFace(face)
	}
}

func (r *Renderer) rasterTriangles(mesh *mobject.Mesh3D, basis cameraBasis, parentAlpha float64, applyBase bool) []rasterTriangle {
	b := mesh.Base()
	if !b.IsVisible || b.Opacity <= 0 || mesh.Wireframe || len(mesh.Vertices) == 0 {
		return nil
	}
	world := make([]geometry.Vec3, len(mesh.Vertices))
	depths := make([]float64, len(mesh.Vertices))
	for i, vertex := range mesh.Vertices {
		world[i] = mesh.Transform3D.Apply(vertex)
		depths[i] = world[i].Sub(basis.position).Dot(basis.forward)
	}
	normals := mesh.VertexNormals
	if mesh.SmoothShading && len(normals) != len(mesh.Vertices) {
		normals = calculateVertexNormals(mesh.Vertices, mesh.Faces)
	}
	worldNormals := make([]geometry.Vec3, len(normals))
	for i, normal := range normals {
		worldNormals[i] = transformNormal(normal, mesh.Transform3D)
	}
	limit := int(math.Ceil(float64(len(mesh.Faces)) * math.Max(0, math.Min(1, b.DrawProgress))))
	if limit > len(mesh.Faces) {
		limit = len(mesh.Faces)
	}
	result := make([]rasterTriangle, 0, limit*2)
	for faceIndex, indices := range mesh.Faces[:limit] {
		if len(indices) < 3 || !validFace(indices, len(world)) {
			continue
		}
		faceNormal := world[indices[1]].Sub(world[indices[0]]).Cross(world[indices[2]].Sub(world[indices[0]])).Normalize()
		if faceNormal.Length() == 0 {
			continue
		}
		center := geometry.Vec3{}
		for _, index := range indices {
			center = center.Add(world[index])
		}
		center = center.Mul(1 / float64(len(indices)))
		if mesh.CullBackfaces && faceNormal.Dot(basis.position.Sub(center)) <= 0 {
			continue
		}
		fill := b.Fill
		if faceIndex < len(mesh.FaceColors) {
			fill = mesh.FaceColors[faceIndex]
		}
		alpha := parentAlpha * b.Opacity * float64(fill.A) / 255
		if alpha <= 0 {
			continue
		}
		for i := 1; i+1 < len(indices); i++ {
			ids := [3]int{indices[0], indices[i], indices[i+1]}
			polygon := make([]clipVertex, 3)
			for j, index := range ids {
				normal := faceNormal
				if mesh.SmoothShading && len(worldNormals) == len(world) {
					normal = worldNormals[index]
				}
				polygon[j] = clipVertex{world: world[index], normal: normal, depth: depths[index]}
			}
			polygon = clipAgainstNear(polygon, basis.near)
			for j := 1; j+1 < len(polygon); j++ {
				clipped := [3]clipVertex{polygon[0], polygon[j], polygon[j+1]}
				var triangle rasterTriangle
				triangle.faceNormal = faceNormal
				triangle.fill = fill
				triangle.alpha = alpha
				triangle.shading = mesh.Shading
				triangle.smoothShading = mesh.SmoothShading
				triangle.emissive = mesh.Emissive
				triangle.specular = mesh.Specular
				triangle.shininess = mesh.Shininess
				valid := true
				for k, vertex := range clipped {
					point := projectClipped(vertex.world, vertex.depth, basis)
					if applyBase {
						point = applyTransform2D(point, b.Transform)
					}
					x, y := r.dc.TransformPoint(point.X, point.Y)
					if math.IsNaN(x) || math.IsNaN(y) || math.IsInf(x, 0) || math.IsInf(y, 0) {
						valid = false
						break
					}
					triangle.vertices[k] = rasterVertex{screen: geometry.V2(x, y), world: vertex.world, normal: vertex.normal, depth: vertex.depth}
					triangle.depth += vertex.depth / 3
				}
				if valid {
					result = append(result, triangle)
				}
			}
		}
	}
	return result
}

func validFace(indices []int, vertexCount int) bool {
	for _, index := range indices {
		if index < 0 || index >= vertexCount {
			return false
		}
	}
	return true
}

func calculateVertexNormals(vertices []geometry.Vec3, faces [][]int) []geometry.Vec3 {
	normals := make([]geometry.Vec3, len(vertices))
	for _, face := range faces {
		if len(face) < 3 || !validFace(face, len(vertices)) {
			continue
		}
		origin := vertices[face[0]]
		for i := 1; i+1 < len(face); i++ {
			normal := vertices[face[i]].Sub(origin).Cross(vertices[face[i+1]].Sub(origin))
			normals[face[0]] = normals[face[0]].Add(normal)
			normals[face[i]] = normals[face[i]].Add(normal)
			normals[face[i+1]] = normals[face[i+1]].Add(normal)
		}
	}
	for i := range normals {
		normals[i] = normals[i].Normalize()
	}
	return normals
}

func transformNormal(normal geometry.Vec3, transform geometry.Transform3D) geometry.Vec3 {
	scale := transform.Scale
	if math.Abs(scale.X) < 1e-12 {
		scale.X = 1
	}
	if math.Abs(scale.Y) < 1e-12 {
		scale.Y = 1
	}
	if math.Abs(scale.Z) < 1e-12 {
		scale.Z = 1
	}
	return geometry.V3(normal.X/scale.X, normal.Y/scale.Y, normal.Z/scale.Z).RotateEuler(transform.Rotation).Normalize()
}

func clipAgainstNear(polygon []clipVertex, near float64) []clipVertex {
	if len(polygon) == 0 {
		return nil
	}
	result := make([]clipVertex, 0, len(polygon)+1)
	previous := polygon[len(polygon)-1]
	previousInside := previous.depth >= near
	for _, current := range polygon {
		currentInside := current.depth >= near
		if currentInside != previousInside {
			delta := current.depth - previous.depth
			if math.Abs(delta) > 1e-12 {
				t := (near - previous.depth) / delta
				result = append(result, clipVertex{
					world:  previous.world.Lerp(current.world, t),
					normal: previous.normal.Lerp(current.normal, t).Normalize(),
					depth:  near,
				})
			}
		}
		if currentInside {
			result = append(result, current)
		}
		previous, previousInside = current, currentInside
	}
	return result
}

func projectClipped(world geometry.Vec3, depth float64, basis cameraBasis) geometry.Vec2 {
	rel := world.Sub(basis.position)
	depth = math.Max(depth, basis.near)
	return geometry.V2(rel.Dot(basis.right)*basis.focal/depth, -rel.Dot(basis.up)*basis.focal/depth)
}

func (r *Renderer) rasterizeTriangles(target *image.RGBA, triangles []rasterTriangle) {
	depthBuffer := make([]float64, target.Bounds().Dx()*target.Bounds().Dy())
	for i := range depthBuffer {
		depthBuffer[i] = math.Inf(1)
	}
	opaque := make([]rasterTriangle, 0, len(triangles))
	transparent := make([]rasterTriangle, 0)
	for _, triangle := range triangles {
		if triangle.alpha >= 0.999 {
			opaque = append(opaque, triangle)
		} else {
			transparent = append(transparent, triangle)
		}
	}
	for i := range opaque {
		r.rasterizeTriangle(target, depthBuffer, &opaque[i], true)
	}
	sort.SliceStable(transparent, func(i, j int) bool { return transparent[i].depth > transparent[j].depth })
	for i := range transparent {
		r.rasterizeTriangle(target, depthBuffer, &transparent[i], false)
	}
}

func edge(a, b, p geometry.Vec2) float64 {
	return (b.X-a.X)*(p.Y-a.Y) - (b.Y-a.Y)*(p.X-a.X)
}

func (r *Renderer) rasterizeTriangle(target *image.RGBA, depthBuffer []float64, triangle *rasterTriangle, writeDepth bool) {
	v := triangle.vertices
	area := edge(v[0].screen, v[1].screen, v[2].screen)
	if math.Abs(area) < 1e-9 {
		return
	}
	if area < 0 {
		v[1], v[2] = v[2], v[1]
		area = -area
	}
	width, height := target.Bounds().Dx(), target.Bounds().Dy()
	minX := max(0, int(math.Floor(math.Min(v[0].screen.X, math.Min(v[1].screen.X, v[2].screen.X)))))
	maxX := min(width-1, int(math.Ceil(math.Max(v[0].screen.X, math.Max(v[1].screen.X, v[2].screen.X)))))
	minY := max(0, int(math.Floor(math.Min(v[0].screen.Y, math.Min(v[1].screen.Y, v[2].screen.Y)))))
	maxY := min(height-1, int(math.Ceil(math.Max(v[0].screen.Y, math.Max(v[1].screen.Y, v[2].screen.Y)))))
	if minX > maxX || minY > maxY {
		return
	}
	invArea := 1 / area
	dx0, dy0 := v[1].screen.Y-v[2].screen.Y, v[2].screen.X-v[1].screen.X
	dx1, dy1 := v[2].screen.Y-v[0].screen.Y, v[0].screen.X-v[2].screen.X
	dx2, dy2 := v[0].screen.Y-v[1].screen.Y, v[1].screen.X-v[0].screen.X
	start := geometry.V2(float64(minX)+0.5, float64(minY)+0.5)
	rowW0 := edge(v[1].screen, v[2].screen, start)
	rowW1 := edge(v[2].screen, v[0].screen, start)
	rowW2 := edge(v[0].screen, v[1].screen, start)
	invDepth := [3]float64{1 / v[0].depth, 1 / v[1].depth, 1 / v[2].depth}
	flatRed, flatGreen, flatBlue := uint8(0), uint8(0), uint8(0)
	var vertexColors [3][3]uint8
	if !triangle.smoothShading {
		center := v[0].world.Add(v[1].world).Add(v[2].world).Mul(1.0 / 3)
		flatRed, flatGreen, flatBlue = r.shadeFragment(triangle, triangle.faceNormal, center, triangle.depth)
	} else {
		for i := range v {
			vertexColors[i][0], vertexColors[i][1], vertexColors[i][2] = r.shadeFragment(triangle, v[i].normal, v[i].world, v[i].depth)
		}
	}
	for y := minY; y <= maxY; y++ {
		w0, w1, w2 := rowW0, rowW1, rowW2
		for x := minX; x <= maxX; x++ {
			if w0 < 0 || w1 < 0 || w2 < 0 {
				w0, w1, w2 = w0+dx0, w1+dx1, w2+dx2
				continue
			}
			b0, b1, b2 := w0*invArea, w1*invArea, w2*invArea
			reciprocalDepth := b0*invDepth[0] + b1*invDepth[1] + b2*invDepth[2]
			if reciprocalDepth <= 0 {
				w0, w1, w2 = w0+dx0, w1+dx1, w2+dx2
				continue
			}
			depth := 1 / reciprocalDepth
			index := y*width + x
			if depth >= depthBuffer[index] {
				w0, w1, w2 = w0+dx0, w1+dx1, w2+dx2
				continue
			}
			red, green, blue := flatRed, flatGreen, flatBlue
			if triangle.smoothShading {
				p0 := b0 * invDepth[0] / reciprocalDepth
				p1 := b1 * invDepth[1] / reciprocalDepth
				p2 := b2 * invDepth[2] / reciprocalDepth
				red = clampByte(p0*float64(vertexColors[0][0]) + p1*float64(vertexColors[1][0]) + p2*float64(vertexColors[2][0]))
				green = clampByte(p0*float64(vertexColors[0][1]) + p1*float64(vertexColors[1][1]) + p2*float64(vertexColors[2][1]))
				blue = clampByte(p0*float64(vertexColors[0][2]) + p1*float64(vertexColors[1][2]) + p2*float64(vertexColors[2][2]))
			}
			blendRGBA(target, x, y, red, green, blue, triangle.alpha)
			if writeDepth {
				depthBuffer[index] = depth
			}
			w0, w1, w2 = w0+dx0, w1+dx1, w2+dx2
		}
		rowW0, rowW1, rowW2 = rowW0+dy0, rowW1+dy1, rowW2+dy2
	}
}

func (r *Renderer) shadeFragment(triangle *rasterTriangle, normal, world geometry.Vec3, depth float64) (uint8, uint8, uint8) {
	red, green, blue := float64(triangle.fill.R), float64(triangle.fill.G), float64(triangle.fill.B)
	if triangle.shading {
		light := r.camera.ThreeD.LightDirection.Normalize()
		diffuse := math.Max(0, normal.Dot(light)) * r.camera.ThreeD.Diffuse
		intensity := triangle.emissive + (1-triangle.emissive)*(r.camera.ThreeD.Ambient+diffuse)
		view := r.camera.ThreeD.Position.Sub(world).Normalize()
		halfVector := light.Add(view).Normalize()
		shininess := math.Max(1, triangle.shininess)
		specular := triangle.specular * math.Pow(math.Max(0, normal.Dot(halfVector)), shininess) * math.Min(1, diffuse*2)
		red = red*intensity + 255*specular
		green = green*intensity + 255*specular
		blue = blue*intensity + 255*specular
	}
	camera := &r.camera.ThreeD
	if camera.FogStrength > 0 && camera.FogFar > camera.FogNear {
		fog := math.Max(0, math.Min(1, (depth-camera.FogNear)/(camera.FogFar-camera.FogNear)))
		fog = fog * fog * (3 - 2*fog) * math.Max(0, math.Min(1, camera.FogStrength))
		red += (float64(camera.FogColor.R) - red) * fog
		green += (float64(camera.FogColor.G) - green) * fog
		blue += (float64(camera.FogColor.B) - blue) * fog
	}
	return clampByte(red), clampByte(green), clampByte(blue)
}

func clampByte(value float64) uint8 {
	return uint8(math.Round(math.Max(0, math.Min(255, value))))
}

func blendRGBA(target *image.RGBA, x, y int, red, green, blue uint8, alpha float64) {
	alpha = math.Max(0, math.Min(1, alpha))
	offset := target.PixOffset(x+target.Bounds().Min.X, y+target.Bounds().Min.Y)
	if alpha >= 0.999 {
		target.Pix[offset], target.Pix[offset+1], target.Pix[offset+2], target.Pix[offset+3] = red, green, blue, 255
		return
	}
	sa := uint32(math.Round(alpha * 255))
	inv := 255 - sa
	target.Pix[offset] = uint8((uint32(red)*sa + uint32(target.Pix[offset])*inv) / 255)
	target.Pix[offset+1] = uint8((uint32(green)*sa + uint32(target.Pix[offset+1])*inv) / 255)
	target.Pix[offset+2] = uint8((uint32(blue)*sa + uint32(target.Pix[offset+2])*inv) / 255)
	target.Pix[offset+3] = uint8(sa + uint32(target.Pix[offset+3])*inv/255)
}

func (r *Renderer) projectMesh(mesh *mobject.Mesh3D, basis cameraBasis, parentAlpha float64, applyBase bool) []projectedFace {
	b := mesh.Base()
	if !b.IsVisible || b.Opacity <= 0 || len(mesh.Vertices) == 0 {
		return nil
	}
	world := make([]geometry.Vec3, len(mesh.Vertices))
	projected := make([]geometry.Vec2, len(mesh.Vertices))
	depths := make([]float64, len(mesh.Vertices))
	visible := make([]bool, len(mesh.Vertices))
	for i, vertex := range mesh.Vertices {
		world[i] = mesh.Transform3D.Apply(vertex)
		projected[i], depths[i], visible[i] = basis.project(world[i])
		if applyBase {
			projected[i] = applyTransform2D(projected[i], b.Transform)
		}
	}
	limit := int(math.Ceil(float64(len(mesh.Faces)) * math.Max(0, math.Min(1, b.DrawProgress))))
	if limit > len(mesh.Faces) {
		limit = len(mesh.Faces)
	}
	result := make([]projectedFace, 0, limit)
	for faceIndex, indices := range mesh.Faces[:limit] {
		if len(indices) < 3 {
			continue
		}
		valid := true
		points := make([]geometry.Vec2, len(indices))
		center := geometry.Vec3{}
		depth := 0.0
		for i, index := range indices {
			if index < 0 || index >= len(world) || !visible[index] {
				valid = false
				break
			}
			points[i] = projected[index]
			center = center.Add(world[index])
			depth += depths[index]
		}
		if !valid {
			continue
		}
		center = center.Mul(1 / float64(len(indices)))
		depth /= float64(len(indices))
		normal := world[indices[1]].Sub(world[indices[0]]).Cross(world[indices[2]].Sub(world[indices[0]])).Normalize()
		if mesh.CullBackfaces && normal.Dot(basis.position.Sub(center)) <= 0 {
			continue
		}

		fill := b.Fill
		if faceIndex < len(mesh.FaceColors) {
			fill = mesh.FaceColors[faceIndex]
		}
		if mesh.Shading {
			light := r.camera.ThreeD.LightDirection.Normalize()
			diffuse := math.Max(0, normal.Dot(light)) * r.camera.ThreeD.Diffuse
			intensity := mesh.Emissive + (1-mesh.Emissive)*(r.camera.ThreeD.Ambient+diffuse)
			fill = shadeColor(fill, intensity)
		}
		result = append(result, projectedFace{
			points: points, depth: depth, fill: fill, stroke: b.Stroke,
			strokeWidth: b.StrokeWidth, alpha: parentAlpha * b.Opacity,
			wireframe: mesh.Wireframe, glowRadius: mesh.GlowRadius, glowColor: mesh.GlowColor,
		})
	}
	return result
}

func applyTransform2D(point geometry.Vec2, transform geometry.Transform) geometry.Vec2 {
	x, y := point.X*transform.Scale.X, point.Y*transform.Scale.Y
	c, s := math.Cos(transform.Rotation), math.Sin(transform.Rotation)
	return geometry.V2(x*c-y*s+transform.Position.X, x*s+y*c+transform.Position.Y)
}

func shadeColor(c color.RGBA, intensity float64) color.RGBA {
	intensity = math.Max(0, math.Min(1.35, intensity))
	channel := func(v uint8) uint8 { return uint8(math.Min(255, math.Round(float64(v)*intensity))) }
	return color.RGBA{R: channel(c.R), G: channel(c.G), B: channel(c.B), A: c.A}
}

func (r *Renderer) drawProjectedFace(face projectedFace) {
	if len(face.points) < 3 {
		return
	}
	drawPath := func() {
		r.dc.MoveTo(face.points[0].X, face.points[0].Y)
		for _, point := range face.points[1:] {
			r.dc.LineTo(point.X, point.Y)
		}
		r.dc.ClosePath()
	}
	if face.glowRadius > 0 && face.glowColor.A > 0 {
		for layer := 3; layer >= 1; layer-- {
			drawPath()
			alpha := face.alpha * (0.045 + 0.035*float64(4-layer))
			r.setColor(withAlpha(face.glowColor, alpha))
			r.dc.SetLineWidth((face.strokeWidth + face.glowRadius*float64(layer)/3) * r.outputScale)
			r.dc.Stroke()
		}
	}
	drawPath()
	if !face.wireframe && face.fill.A > 0 {
		r.setColor(withAlpha(face.fill, face.alpha))
		r.dc.FillPreserve()
	}
	if face.stroke.A > 0 && face.strokeWidth > 0 {
		r.setColor(withAlpha(face.stroke, face.alpha))
		r.dc.SetLineWidth(face.strokeWidth * r.outputScale)
		r.dc.Stroke()
	} else if !face.wireframe && face.fill.A > 0 {
		// Adjacent polygons are rasterized independently. A thin same-color
		// stroke overlaps their antialiased edges so the background cannot leak
		// through as an accidental wireframe grid.
		r.setColor(withAlpha(face.fill, face.alpha))
		r.dc.SetLineWidth(0.9 * r.outputScale)
		r.dc.Stroke()
	} else {
		r.dc.ClearPath()
	}
}
