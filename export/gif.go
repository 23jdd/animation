package export

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/color/palette"
	"image/gif"
	"io"
	"math"
	"os"
	"sort"
	"sync"

	"github.com/23dd/animation/renderer"
	"github.com/23dd/animation/scene"
	"github.com/23dd/animation/timeline"
)

func GIF(path string, s *scene.Scene, t *timeline.Timeline, r renderer.Renderer, fps float64) error {
	return GIFWithProgress(path, s, t, r, fps, nil)
}

// GIFWithProgress encodes one frame at a time; memory does not grow with duration.
func GIFWithProgress(path string, s *scene.Scene, t *timeline.Timeline, r renderer.Renderer, fps float64, progress func(int, int)) error {
	if fps <= 0 {
		fps = 30
	}
	if math.IsNaN(fps) || math.IsInf(fps, 0) || fps > 100 {
		return fmt.Errorf("GIF fps must be finite and at most 100")
	}
	duration := t.Duration()
	frames := int(duration*fps) + 1
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	var encoded bytes.Buffer
	var dst *image.Paletted
	lookupOnce.Do(buildLookup)
	for i := 0; i < frames; i++ {
		t.Seek(float64(i) / fps)
		s.Update(t.Time, 1/fps)
		src := r.Render(s)
		if dst == nil {
			dst = image.NewPaletted(src.Bounds(), palette.Plan9)
		}
		if src.Bounds() != dst.Bounds() {
			return fmt.Errorf("GIF frame dimensions changed")
		}
		quantizeGIF(dst, src)
		// Rounded cumulative timestamps avoid speeding up 30/60 fps animations.
		delay := int(math.Round(float64(i+1)*100/fps) - math.Round(float64(i)*100/fps))
		encoded.Reset()
		// An explicit non-palette Config makes each frame use a local color table.
		err = gif.EncodeAll(&encoded, &gif.GIF{Image: []*image.Paletted{dst}, Delay: []int{delay},
			Disposal: []byte{gif.DisposalNone}, LoopCount: -1,
			Config: image.Config{Width: dst.Bounds().Dx(), Height: dst.Bounds().Dy()}})
		if err != nil {
			return err
		}
		data := encoded.Bytes()
		if len(data) < 14 || data[10]&0x80 != 0 {
			return fmt.Errorf("unexpected GIF frame header")
		}
		if i == 0 {
			if _, err = w.Write(data[:13]); err != nil {
				return err
			}
			if _, err = io.WriteString(w, "\x21\xff\x0bNETSCAPE2.0\x03\x01\x00\x00\x00"); err != nil {
				return err
			}
		}
		// Keep frame extensions, local palette and LZW data; omit header/trailer.
		if _, err = w.Write(data[13 : len(data)-1]); err != nil {
			return err
		}
		if progress != nil {
			progress(i+1, frames)
		}
	}
	if err = w.WriteByte(0x3b); err != nil {
		return err
	}
	if err = w.Flush(); err != nil {
		return err
	}
	return f.Close()
}

var lookupOnce sync.Once
var gifLookup [32768]uint8
var gifColors [256][3]int

func buildLookup() {
	for i, c := range palette.Plan9 {
		r, g, b, _ := c.RGBA()
		gifColors[i] = [3]int{int(r >> 8), int(g >> 8), int(b >> 8)}
	}
	for key := range gifLookup {
		r, g, b := ((key>>10)&31)*8+4, ((key>>5)&31)*8+4, (key&31)*8+4
		best, distance := 0, math.MaxInt
		for i, c := range gifColors[:240] {
			dr, dg, db := r-c[0], g-c[1], b-c[2]
			d := dr*dr + dg*dg + db*db
			if d < distance {
				best, distance = i, d
			}
		}
		gifLookup[key] = uint8(best)
	}
}

// Floyd–Steinberg error diffusion with constant-time 5-bit palette lookup.
func quantizeGIF(dst *image.Paletted, src image.Image) {
	colors, lookup := gifColors, gifLookup
	// Reserve 16 entries for the frame's dominant exact colors. This prevents
	// dithering on flat backgrounds while retaining a general-purpose palette.
	counts := make(map[uint32]int)
	step := max(1, int(math.Sqrt(float64(src.Bounds().Dx()*src.Bounds().Dy())/16384)))
	for y := src.Bounds().Min.Y; y < src.Bounds().Max.Y; y += step {
		for x := src.Bounds().Min.X; x < src.Bounds().Max.X; x += step {
			r, g, b, _ := src.At(x, y).RGBA()
			counts[(r>>8)<<16|(g>>8)<<8|b>>8]++
		}
	}
	keys := make([]uint32, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if counts[keys[i]] == counts[keys[j]] {
			return keys[i] < keys[j]
		}
		return counts[keys[i]] > counts[keys[j]]
	})
	dst.Palette = append(color.Palette(nil), palette.Plan9...)
	for i, k := range keys[:min(16, len(keys))] {
		colors[240+i] = [3]int{int(k >> 16), int(k >> 8 & 255), int(k & 255)}
		dst.Palette[240+i] = color.RGBA{uint8(k >> 16), uint8(k >> 8), uint8(k), 255}
	}
	for key, index := range lookup {
		r, g, b := ((key>>10)&31)*8+4, ((key>>5)&31)*8+4, (key&31)*8+4
		c := colors[index]
		dr, dg, db := r-c[0], g-c[1], b-c[2]
		distance := dr*dr + dg*dg + db*db
		for i, c := range colors[240:] {
			dr, dg, db = r-c[0], g-c[1], b-c[2]
			d := dr*dr + dg*dg + db*db
			if d < distance {
				lookup[key] = uint8(240 + i)
				distance = d
			}
		}
	}
	// Preserve dominant colors exactly, even near a lookup-cell boundary.
	for i := min(16, len(keys)) - 1; i >= 0; i-- {
		k := keys[i]
		lookup[((k>>19)&31)<<10|((k>>11)&31)<<5|((k>>3)&31)] = uint8(240 + i)
	}
	w := src.Bounds().Dx()
	current, next := make([][3]int, w+2), make([][3]int, w+2)
	rgba, direct := src.(*image.RGBA)
	for y := src.Bounds().Min.Y; y < src.Bounds().Max.Y; y++ {
		for x := 0; x < w; x++ {
			var rgb [3]int
			if direct {
				o := rgba.PixOffset(x+src.Bounds().Min.X, y)
				rgb = [3]int{int(rgba.Pix[o]), int(rgba.Pix[o+1]), int(rgba.Pix[o+2])}
			} else {
				r, g, b, _ := src.At(x+src.Bounds().Min.X, y).RGBA()
				rgb = [3]int{int(r >> 8), int(g >> 8), int(b >> 8)}
			}
			for c := range rgb {
				rgb[c] = max(0, min(255, rgb[c]+current[x+1][c]/16))
			}
			index := lookup[(rgb[0]>>3)<<10|(rgb[1]>>3)<<5|rgb[2]>>3]
			dst.Pix[(y-src.Bounds().Min.Y)*dst.Stride+x] = index
			for c := range rgb {
				e := rgb[c] - colors[index][c]
				current[x+2][c] += e * 7
				next[x][c] += e * 3
				next[x+1][c] += e * 5
				next[x+2][c] += e
			}
		}
		current, next = next, current
		clear(next)
	}
}
