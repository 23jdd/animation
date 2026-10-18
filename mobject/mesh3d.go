package mobject

import (
	"image/color"
	"math"

	"github.com/23dd/animation/geometry"
)

// Mesh3D is a polygon mesh projected by the renderer's 3D camera.
type Mesh3D struct {
	BaseMobject
	Transform3D   geometry.Transform3D
	Vertices      []geometry.Vec3
	VertexNormals []geometry.Vec3
	Faces         [][]int
	FaceColors    []color.RGBA
	CullBackfaces bool
	Wireframe     bool
	Shading       bool
	SmoothShading bool
	Emissive      float64
	Specular      float64
	Shininess     float64
	GlowRadius    float64
	GlowColor     color.RGBA
}

func NewMesh3D(vertices []geometry.Vec3, faces [][]int) *Mesh3D {
	clonedFaces := make([][]int, len(faces))
	for i := range faces {
		clonedFaces[i] = append([]int(nil), faces[i]...)
	}
	m := &Mesh3D{
		BaseMobject: NewBase(),
		Transform3D: geometry.NewTransform3D(),
		Vertices:    append([]geometry.Vec3(nil), vertices...),
		Faces:       clonedFaces,
		Shading:     true,
		Emissive:    0.08,
		Specular:    0.12,
		Shininess:   32,
		GlowColor:   color.RGBA{255, 255, 255, 255},
	}
	return m
}

func (m *Mesh3D) Update(t, dt float64) error { return Update(m, t, dt) }

func NewBox3D(width, height, depth float64) *Mesh3D {
	x, y, z := width/2, height/2, depth/2
	vertices := []geometry.Vec3{
		geometry.V3(-x, -y, -z), geometry.V3(x, -y, -z), geometry.V3(x, y, -z), geometry.V3(-x, y, -z),
		geometry.V3(-x, -y, z), geometry.V3(x, -y, z), geometry.V3(x, y, z), geometry.V3(-x, y, z),
	}
	faces := [][]int{
		{4, 5, 6, 7}, {1, 0, 3, 2},
		{0, 4, 7, 3}, {5, 1, 2, 6},
		{3, 7, 6, 2}, {0, 1, 5, 4},
	}
	m := NewMesh3D(vertices, faces)
	m.CullBackfaces = true
	return m
}

func NewSphere3D(radius float64, segments, rings int) *Mesh3D {
	segments = max(3, segments)
	rings = max(2, rings)
	vertices := make([]geometry.Vec3, 0, (rings+1)*segments)
	for i := 0; i <= rings; i++ {
		phi := -math.Pi/2 + math.Pi*float64(i)/float64(rings)
		cp, sp := math.Cos(phi), math.Sin(phi)
		for j := 0; j < segments; j++ {
			theta := 2 * math.Pi * float64(j) / float64(segments)
			vertices = append(vertices, geometry.V3(radius*cp*math.Cos(theta), radius*sp, radius*cp*math.Sin(theta)))
		}
	}
	faces := make([][]int, 0, rings*segments)
	for i := 0; i < rings; i++ {
		for j := 0; j < segments; j++ {
			next := (j + 1) % segments
			a, b := i*segments+j, i*segments+next
			c, d := (i+1)*segments+j, (i+1)*segments+next
			faces = append(faces, []int{a, c, d, b})
		}
	}
	m := NewMesh3D(vertices, faces)
	m.CullBackfaces = true
	m.SmoothShading = true
	m.VertexNormals = make([]geometry.Vec3, len(vertices))
	for i, vertex := range vertices {
		m.VertexNormals[i] = vertex.Normalize()
	}
	return m
}

func NewTorus3D(majorRadius, minorRadius float64, segments, sides int) *Mesh3D {
	segments = max(3, segments)
	sides = max(3, sides)
	vertices := make([]geometry.Vec3, 0, segments*sides)
	normals := make([]geometry.Vec3, 0, segments*sides)
	for i := 0; i < segments; i++ {
		u := 2 * math.Pi * float64(i) / float64(segments)
		cu, su := math.Cos(u), math.Sin(u)
		for j := 0; j < sides; j++ {
			v := 2 * math.Pi * float64(j) / float64(sides)
			cv, sv := math.Cos(v), math.Sin(v)
			ring := majorRadius + minorRadius*cv
			vertices = append(vertices, geometry.V3(ring*cu, minorRadius*sv, ring*su))
			normals = append(normals, geometry.V3(cu*cv, sv, su*cv))
		}
	}
	faces := make([][]int, 0, segments*sides)
	for i := 0; i < segments; i++ {
		ni := (i + 1) % segments
		for j := 0; j < sides; j++ {
			nj := (j + 1) % sides
			a, b := i*sides+j, i*sides+nj
			c, d := ni*sides+j, ni*sides+nj
			faces = append(faces, []int{a, c, d, b})
		}
	}
	m := NewMesh3D(vertices, faces)
	m.CullBackfaces = true
	m.SmoothShading = true
	m.VertexNormals = normals
	return m
}
