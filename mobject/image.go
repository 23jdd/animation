package mobject

import "image"

// ImageFit determines how a bitmap fits its logical display box.
type ImageFit string

const (
	ImageContain ImageFit = "contain"
	ImageCover   ImageFit = "cover"
	ImageStretch ImageFit = "stretch"
)

// Image holds decoded pixels, independently of any renderer or filesystem.
// Pixels may be shared by several objects and are treated as immutable.
type Image struct {
	BaseMobject
	Pixels        image.Image
	Width, Height float64
	Fit           ImageFit
}

func NewImage(pixels image.Image) *Image {
	m := &Image{BaseMobject: NewBase(), Pixels: pixels, Fit: ImageContain}
	if pixels != nil {
		m.Width = float64(pixels.Bounds().Dx())
		m.Height = float64(pixels.Bounds().Dy())
	}
	return m
}

func (m *Image) Update(t, dt float64) error { return Update(m, t, dt) }
