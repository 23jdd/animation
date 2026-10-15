package gg

import (
	"errors"
	"image"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// fallbackFace uses primary whenever it contains a rune and falls back one
// rune at a time. This mirrors the behavior users expect from IDE font stacks.
type fallbackFace struct {
	primary  font.Face
	fallback font.Face
}

func (f *fallbackFace) Close() error {
	return errors.Join(f.primary.Close(), f.fallback.Close())
}

func (f *fallbackFace) Glyph(dot fixed.Point26_6, r rune) (image.Rectangle, image.Image, image.Point, fixed.Int26_6, bool) {
	return f.faceFor(r).Glyph(dot, r)
}

func (f *fallbackFace) GlyphBounds(r rune) (fixed.Rectangle26_6, fixed.Int26_6, bool) {
	return f.faceFor(r).GlyphBounds(r)
}

func (f *fallbackFace) GlyphAdvance(r rune) (fixed.Int26_6, bool) {
	return f.faceFor(r).GlyphAdvance(r)
}

func (f *fallbackFace) Kern(r0, r1 rune) fixed.Int26_6 {
	primary0 := supportsRune(f.primary, r0)
	primary1 := supportsRune(f.primary, r1)
	if primary0 && primary1 {
		return f.primary.Kern(r0, r1)
	}
	if !primary0 && !primary1 {
		return f.fallback.Kern(r0, r1)
	}
	return 0
}

func (f *fallbackFace) Metrics() font.Metrics { return f.primary.Metrics() }

func (f *fallbackFace) faceFor(r rune) font.Face {
	if supportsRune(f.primary, r) {
		return f.primary
	}
	return f.fallback
}

func supportsRune(face font.Face, r rune) bool {
	_, ok := face.GlyphAdvance(r)
	return ok
}
