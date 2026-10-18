package script

import (
	"fmt"
	"image"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/23dd/animation/mobject"
	"github.com/dop251/goja"
	_ "golang.org/x/image/webp"
)

const maxImageBytes = 64 << 20
const maxImagePixels = 64_000_000

func (r *Runtime) loadImage(path string) (image.Image, error) {
	if pixels := r.images[path]; pixels != nil {
		return pixels, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("image path must point to a regular file")
	}
	if info.Size() > maxImageBytes {
		return nil, fmt.Errorf("image file exceeds the 64 MiB limit")
	}
	config, format, err := image.DecodeConfig(io.LimitReader(f, maxImageBytes))
	if err != nil {
		return nil, fmt.Errorf("decode image header (PNG, JPEG, WebP or GIF required): %w", err)
	}
	if format != "png" && format != "jpeg" && format != "webp" && format != "gif" {
		return nil, fmt.Errorf("unsupported image format %q", format)
	}
	if config.Width <= 0 || config.Height <= 0 || config.Width > maxImagePixels/config.Height {
		return nil, fmt.Errorf("image dimensions must be positive and at most 64 million pixels")
	}
	if _, err = f.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	pixels, _, err := image.Decode(io.LimitReader(f, maxImageBytes))
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}
	// GIF frames may occupy only a subrectangle of their logical canvas.
	// Preserve that canvas and the frame offset instead of centering the tile.
	canvas := image.Rect(0, 0, config.Width, config.Height)
	if format == "gif" && pixels.Bounds() != canvas {
		frame := image.NewRGBA(canvas)
		draw.Draw(frame, pixels.Bounds(), pixels, pixels.Bounds().Min, draw.Src)
		pixels = frame
	}
	if r.images == nil {
		r.images = make(map[string]image.Image)
	}
	r.images[path] = pixels
	return pixels, nil
}

func (r *Runtime) jsImage(c goja.FunctionCall) goja.Value {
	if !isDefined(c.Argument(0)) {
		r.fail("image source must be a non-empty local file path")
		return goja.Undefined()
	}
	source, ok := c.Argument(0).Export().(string)
	if !ok || strings.TrimSpace(source) == "" {
		r.fail("image source must be a non-empty local file path")
		return goja.Undefined()
	}
	if strings.Contains(source, "://") || strings.HasPrefix(source, "data:") {
		r.fail("image supports local file paths, not URLs or data URLs")
		return goja.Undefined()
	}
	if v := c.Argument(1); isDefined(v) {
		if o, ok := v.(*goja.Object); !ok || o.ClassName() != "Object" {
			r.fail("image options must be an object")
			return goja.Undefined()
		}
	}
	o := object(c.Argument(1), r.VM)
	fit := mobject.ImageContain
	if v := o.Get("fit"); isDefined(v) {
		name, ok := v.Export().(string)
		if !ok || (name != "contain" && name != "cover" && name != "stretch") {
			r.fail("image fit must be contain, cover or stretch")
			return goja.Undefined()
		}
		fit = mobject.ImageFit(name)
	}
	width := r.finiteNumber(o.Get("width"), 0, "image width")
	height := r.finiteNumber(o.Get("height"), 0, "image height")
	if (isDefined(o.Get("width")) && width <= 0) || (isDefined(o.Get("height")) && height <= 0) {
		r.fail("image width and height must be positive")
	}
	if r.callErr != nil {
		return goja.Undefined()
	}
	path := source
	if !filepath.IsAbs(path) {
		path = filepath.Join(r.BaseDir, path)
	}
	path = filepath.Clean(path)
	pixels, err := r.loadImage(path)
	if err != nil {
		r.fail(fmt.Sprintf("load image %q: %v", source, err))
		return goja.Undefined()
	}
	m := mobject.NewImage(pixels)
	naturalWidth, naturalHeight := m.Width, m.Height
	switch {
	case width == 0 && height == 0:
		width, height = naturalWidth, naturalHeight
	case width == 0:
		width = height * naturalWidth / naturalHeight
	case height == 0:
		height = width * naturalHeight / naturalWidth
	}
	if math.IsInf(width, 0) || math.IsInf(height, 0) || width <= 0 || height <= 0 {
		r.fail("image display dimensions are out of range")
		return goja.Undefined()
	}
	m.Width, m.Height, m.Fit = width, height, fit
	style(m, o)
	wrapped := r.wrap(m).ToObject(r.VM)
	for name, value := range map[string]interface{}{
		"width": width, "height": height, "naturalWidth": naturalWidth,
		"naturalHeight": naturalHeight, "source": path, "fit": string(fit),
	} {
		if err := wrapped.DefineDataProperty(name, r.VM.ToValue(value), goja.FLAG_FALSE, goja.FLAG_FALSE, goja.FLAG_TRUE); err != nil {
			r.fail("define image property " + name + ": " + err.Error())
		}
	}
	return wrapped
}
