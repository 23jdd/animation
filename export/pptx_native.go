package export

// Native export samples object state, not pixels. Keeping this separate from
// the renderer lets arbitrary deterministic updaters produce editable keyframes.
import (
	"errors"
	"fmt"
	"image/color"
	"io"
	"math"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/23dd/animation/geometry"
	"github.com/23dd/animation/mobject"
	"github.com/23dd/animation/scene"
	"github.com/23dd/animation/timeline"
)

const maxNativeStates = 1_000_000

type nativeState struct {
	x, y, w, h, rotation, alpha float64
}

type nativeTrack struct {
	id     uint64
	spec   nativeSpec
	states []nativeState
}

type nativeDeck struct {
	width, height               int32
	logicalWidth, logicalHeight int
	duration                    int
	times                       []int
	background                  color.RGBA
	tracks                      []nativeTrack
}

// PPTXNative exports one slide containing editable 2D objects and automatic
// PresentationML animation. fps (1..120) controls keyframe sampling, not video
// frames. Unsupported state changes return an error instead of rasterizing.
func PPTXNative(path string, s *scene.Scene, t *timeline.Timeline, fps float64) error {
	if !strings.EqualFold(filepath.Ext(path), ".pptx") {
		return errors.New("PPTX output must have a .pptx extension")
	}
	if s == nil || s.Camera == nil || t == nil {
		return errors.New("native PPTX requires a scene, camera and timeline")
	}
	d, err := sampleNativeDeck(s, t, fps)
	if err != nil {
		return fmt.Errorf("native PPTX: %w (use frame export without -native for unsupported scenes)", err)
	}
	slide, err := d.slideXML()
	if err != nil {
		return fmt.Errorf("native PPTX: %w", err)
	}
	return savePPTX(path, func(out io.Writer) error {
		p, err := newPPTXArchive(out, d.width, d.height, strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)), 1)
		if err != nil {
			return err
		}
		err = p.addNativeSlide(slide, float64(d.duration)/1000)
		return errors.Join(err, p.zip.Close())
	})
}

func nativeSampleTimes(t *timeline.Timeline, fps float64) ([]int, error) {
	duration := t.Duration()
	if !nativeFinite(fps) || fps < 1 || fps > 120 {
		return nil, errors.New("-fps must be finite and between 1 and 120 for native PPTX")
	}
	if !nativeFinite(duration) || duration < 0 || duration > 300 {
		return nil, errors.New("native timeline duration must be between 0 and 300 seconds")
	}
	end := int(math.Round(duration * 1000))
	if duration > 0 && end == 0 {
		return nil, errors.New("native animation must last at least one millisecond")
	}
	if duration*fps > 18000 {
		return nil, errors.New("native export exceeds 18000 samples; reduce -fps")
	}
	times := []int{0, end}
	for i := 1; float64(i)/fps < duration; i++ {
		times = append(times, int(math.Round(float64(i)*1000/fps)))
	}
	// Retain top-level boundaries even when they fall between regular samples.
	for _, track := range t.Tracks {
		if track.Animation == nil || (reflect.ValueOf(track.Animation).Kind() == reflect.Pointer && reflect.ValueOf(track.Animation).IsNil()) {
			return nil, errors.New("timeline contains a nil animation")
		}
		for _, at := range []float64{track.Start, track.Start + track.Animation.Duration()} {
			if !nativeFinite(at) || at < 0 || at > duration {
				return nil, errors.New("timeline contains an invalid track boundary")
			}
			times = append(times, int(math.Round(at*1000)))
		}
	}
	sort.Ints(times)
	unique := times[:0]
	for _, at := range times {
		if len(unique) == 0 || at != unique[len(unique)-1] {
			unique = append(unique, at)
		}
	}
	if len(unique) > 20000 {
		return nil, errors.New("native export exceeds 20000 sample boundaries")
	}
	return unique, nil
}

func sampleNativeDeck(s *scene.Scene, t *timeline.Timeline, fps float64) (d *nativeDeck, result error) {
	times, err := nativeSampleTimes(t, fps)
	if err != nil {
		return nil, err
	}
	w, h, err := pptxSlideSize(s.Camera.Width, s.Camera.Height)
	if err != nil {
		return nil, err
	}
	previous := t.Time
	defer func() {
		t.Seek(previous)
		if err := s.Update(t.Time, 0); err != nil {
			result = errors.Join(result, fmt.Errorf("restore scene: %w", err))
		}
	}()
	d = &nativeDeck{width: w, height: h, logicalWidth: s.Camera.Width, logicalHeight: s.Camera.Height, duration: times[len(times)-1], times: times}
	builder := newNativeBuilder()
	for index, ms := range times {
		t.Seek(math.Min(t.Duration(), float64(ms)/1000))
		if err := s.Update(t.Time, 0); err != nil {
			return nil, fmt.Errorf("update at %gs: %w", t.Time, err)
		}
		if s.Camera == nil {
			return nil, errors.New("scene camera was removed during export")
		}
		if index == 0 {
			d.background = s.Camera.Background
		} else if s.Camera.Background != d.background {
			return nil, errors.New("animated camera background is not supported")
		}
		if !nativeFinite(s.Camera.Zoom) || s.Camera.Zoom <= 0 {
			return nil, errors.New("camera zoom must be positive and finite")
		}
		cam := s.Camera
		if cam.Width != d.logicalWidth || cam.Height != d.logicalHeight {
			return nil, errors.New("changing canvas dimensions during native export is not supported")
		}
		matrix := geometry.Translate(geometry.V2(float64(cam.Width)/2-cam.Position.X*cam.Zoom, float64(cam.Height)/2-cam.Position.Y*cam.Zoom)).Mul(geometry.Scale(geometry.V2(cam.Zoom, cam.Zoom)))
		current, err := builder.snapshot(s.Snapshot(), matrix, 1, 0)
		if err != nil {
			return nil, fmt.Errorf("at %gs: %w", t.Time, err)
		}
		if index == 0 {
			if len(current) > 500 || len(current)*len(times) > maxNativeStates {
				return nil, errors.New("native export exceeds 500 objects or one million object samples")
			}
			for _, track := range current {
				state, spec := track.states[0], track.spec
				if spec.family != "" && math.Abs(state.w/spec.width-state.h/spec.height) > 1e-7 {
					return nil, errors.New("native text must start with uniform scale")
				}
			}
			d.tracks = current
		} else {
			if len(current) != len(d.tracks) {
				return nil, errors.New("adding/removing objects during native export is not supported")
			}
			for i, track := range current {
				if track.id != d.tracks[i].id || !reflect.DeepEqual(track.spec, d.tracks[i].spec) {
					return nil, fmt.Errorf("object %d: geometry, text, fill, stroke or drawing progress changes are not supported", i+1)
				}
				d.tracks[i].states = append(d.tracks[i].states, track.states[0])
			}
		}
	}
	return d, nil
}

func (b *nativeBuilder) snapshot(objects []mobject.Mobject, parent geometry.Matrix, alpha, angle float64) ([]nativeTrack, error) {
	var out []nativeTrack
	seen := make(map[uint64]bool)
	var walk func([]mobject.Mobject, geometry.Matrix, float64, float64, int) error
	walk = func(items []mobject.Mobject, parent geometry.Matrix, alpha, angle float64, depth int) error {
		if depth > 64 {
			return errors.New("group hierarchy exceeds 64 levels")
		}
		for _, obj := range items {
			if obj == nil || (reflect.ValueOf(obj).Kind() == reflect.Pointer && reflect.ValueOf(obj).IsNil()) || seen[obj.ID()] {
				return errors.New("nil, duplicate or cyclic scene object")
			}
			seen[obj.ID()] = true
			base := obj.Base()
			if !nativeFinite(base.Transform.Scale.X) || !nativeFinite(base.Transform.Scale.Y) || base.Transform.Scale.X <= 0 || base.Transform.Scale.Y <= 0 {
				return errors.New("native objects require positive finite scales")
			}
			m := parent.Mul(base.Transform.Matrix())
			a := alpha * base.Opacity
			if !base.IsVisible {
				a = 0
			}
			rotation := angle + base.Transform.Rotation
			if !nativeFinite(a) || a < 0 || a > 1 {
				return errors.New("object opacity must be finite and between 0 and 1")
			}
			if math.Abs(base.DrawProgress-1) > 1e-9 {
				return errors.New("create/uncreate/draw effects are not supported in native mode")
			}
			if group, ok := obj.(*mobject.Group); ok {
				if err := walk(group.Children, m, a, rotation, depth+1); err != nil {
					return err
				}
				continue
			}
			spec, err := b.spec(obj)
			if err != nil {
				return err
			}
			sx, sy := math.Hypot(m.A, m.B), math.Hypot(m.C, m.D)
			if sx <= 1e-9 || sy <= 1e-9 || m.A*m.D-m.B*m.C <= 0 || math.Abs(m.A*m.C+m.B*m.D) > 1e-7*sx*sy {
				return errors.New("zero/negative scale, reflections and sheared groups are not supported")
			}
			center := m.Apply(geometry.V2(spec.left+spec.width/2, spec.top+spec.height/2))
			r := math.Atan2(m.B, m.A)
			r += math.Round((rotation-r)/(2*math.Pi)) * 2 * math.Pi
			state := nativeState{center.X, center.Y, spec.width * sx, spec.height * sy, r * 180 / math.Pi, a}
			if math.Abs(state.rotation*60000) > 2147483647 {
				return errors.New("rotation exceeds OOXML angle limits")
			}
			for _, v := range []float64{state.x, state.y, state.w, state.h, state.rotation} {
				if !nativeFinite(v) || math.Abs(v) > 1e8 {
					return errors.New("non-finite or excessive object transform")
				}
			}
			out = append(out, nativeTrack{id: obj.ID(), spec: spec, states: []nativeState{state}})
			if len(out) > 500 {
				return errors.New("native export supports at most 500 objects")
			}
		}
		return nil
	}
	if err := walk(objects, parent, alpha, angle, 0); err != nil {
		return nil, err
	}
	return out, nil
}

func nativeFinite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

func (d *nativeDeck) canvasWidth() float64  { return float64(d.logicalWidth) }
func (d *nativeDeck) canvasHeight() float64 { return float64(d.logicalHeight) }
