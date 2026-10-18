package export

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
)

type nativeFade struct {
	start, end int
	in         bool
}

// A fade filter applies to the whole shape, including its text and outline.
// Arbitrary partial-opacity changes cannot be faithfully encoded by this
// filter, so reject them rather than emit an ineffective style.opacity track.
func nativeFades(states []nativeState, times []int) (float64, []nativeFade, error) {
	peak := 0.0
	for _, state := range states {
		peak = math.Max(peak, state.alpha)
	}
	if peak == 0 {
		return 0, nil, nil
	}
	level := func(i int) float64 { return states[i].alpha / peak }
	endpoint := func(v float64) bool { return v < 1e-6 || v > 1-1e-6 }
	if !endpoint(level(0)) {
		return 0, nil, errors.New("partial-opacity animation is not supported; use complete fadeIn/fadeOut")
	}
	var fades []nativeFade
	for i := 0; i+1 < len(states); {
		if math.Abs(level(i+1)-level(i)) < 1e-9 {
			i++
			continue
		}
		start := i
		in := level(i+1) > level(i)
		for i+1 < len(states) {
			delta := level(i+1) - level(i)
			if (in && delta < -1e-7) || (!in && delta > 1e-7) {
				return 0, nil, errors.New("non-monotonic or partial-opacity fade is not supported")
			}
			i++
			if endpoint(level(i)) {
				break
			}
		}
		if !endpoint(level(i)) || math.Abs(level(start)-level(i)) < .99 {
			return 0, nil, errors.New("partial-opacity animation is not supported; use complete fadeIn/fadeOut")
		}
		fades = append(fades, nativeFade{times[start], times[i], in})
	}
	return peak, fades, nil
}

type nativeTiming struct {
	nextID int
	buf    bytes.Buffer
}

func (w *nativeTiming) id() int { w.nextID++; return w.nextID }

func (w *nativeTiming) behavior(shape int, duration, delay int, attribute string) string {
	return fmt.Sprintf(`<p:cBhvr additive="base"><p:cTn id="%d" dur="%d" fill="hold"><p:stCondLst><p:cond delay="%d"/></p:stCondLst></p:cTn><p:tgtEl><p:spTgt spid="%d"/></p:tgtEl><p:attrNameLst><p:attrName>%s</p:attrName></p:attrNameLst></p:cBhvr>`, w.id(), max(1, duration), delay, shape, attribute)
}

func (w *nativeTiming) visibility(shape, delay int, visible bool) {
	value := "hidden"
	if visible {
		value = "visible"
	}
	behavior := strings.Replace(w.behavior(shape, 1, delay, "style.visibility"), ` additive="base"`, "", 1)
	w.buf.WriteString(`<p:set>` + behavior + `<p:to><p:strVal val="` + value + `"/></p:to></p:set>`)
}

// Reduce collinear samples while retaining ease curves and wait plateaus.
func nativeKeyIndices(values []float64, times []int, tolerance float64) []int {
	if len(values) < 2 {
		return []int{0}
	}
	indices := []int{0}
	var split func(int, int)
	split = func(left, right int) {
		worst, at := tolerance, -1
		for i := left + 1; i < right; i++ {
			p := float64(times[i]-times[left]) / float64(times[right]-times[left])
			error := math.Abs(values[i] - (values[left] + (values[right]-values[left])*p))
			if error > worst {
				worst, at = error, i
			}
		}
		if at >= 0 {
			split(left, at)
			indices = append(indices, at)
			split(at, right)
		}
	}
	split(0, len(values)-1)
	return append(indices, len(values)-1)
}

func (w *nativeTiming) property(shape, duration int, name string, values []float64, times []int) {
	changed := false
	for _, value := range values {
		if math.Abs(value-values[0]) > 1e-8 {
			changed = true
			break
		}
	}
	if !changed || duration == 0 {
		return
	}
	w.buf.WriteString(`<p:anim calcmode="lin" valueType="num">` + w.behavior(shape, duration, 0, name) + `<p:tavLst>`)
	indices := nativeKeyIndices(values, times, 1e-6)
	last := -1
	for j, i := range indices {
		tm := int(math.Round(float64(times[i]) / float64(duration) * 100000))
		if j+1 < len(indices) && tm == int(math.Round(float64(times[indices[j+1]])/float64(duration)*100000)) {
			continue
		}
		if tm <= last {
			continue
		}
		last = tm
		fmt.Fprintf(&w.buf, `<p:tav tm="%d"><p:val><p:fltVal val="%.9g"/></p:val></p:tav>`, tm, values[i])
	}
	w.buf.WriteString(`</p:tavLst></p:anim>`)
}

func (w *nativeTiming) rotations(shape int, values []float64, times []int) {
	indices := nativeKeyIndices(values, times, .02)
	for i := 1; i < len(indices); i++ {
		left, right := indices[i-1], indices[i]
		if math.Abs(values[right]-values[left]) < 1e-7 {
			continue
		}
		fmt.Fprintf(&w.buf, `<p:animRot from="%d" to="%d">%s</p:animRot>`, int64(math.Round(values[left]*60000)), int64(math.Round(values[right]*60000)), w.behavior(shape, times[right]-times[left], times[left], "r"))
	}
}

func (w *nativeTiming) scales(shape int, states []nativeState, times []int) error {
	x, y := make([]float64, len(states)), make([]float64, len(states))
	for i, state := range states {
		x[i], y[i] = state.w/states[0].w, state.h/states[0].h
		if x[i]*100000 > 2147483647 || y[i]*100000 > 2147483647 {
			return errors.New("scale animation exceeds OOXML limits")
		}
	}
	indices := append(nativeKeyIndices(x, times, .0002), nativeKeyIndices(y, times, .0002)...)
	sort.Ints(indices)
	previous := indices[0]
	for _, current := range indices[1:] {
		if current == previous {
			continue
		}
		if math.Abs(x[current]-x[previous]) > 1e-7 || math.Abs(y[current]-y[previous]) > 1e-7 {
			// Unlike resizing a text box, animScale also scales its contents.
			behavior := w.behavior(shape, times[current]-times[previous], times[previous], "ScaleX")
			behavior = strings.Replace(behavior, "</p:attrNameLst>", "<p:attrName>ScaleY</p:attrName></p:attrNameLst>", 1)
			fmt.Fprintf(&w.buf, `<p:animScale>%s<p:from x="%d" y="%d"/><p:to x="%d" y="%d"/></p:animScale>`, behavior, int64(math.Round(x[previous]*100000)), int64(math.Round(y[previous]*100000)), int64(math.Round(x[current]*100000)), int64(math.Round(y[current]*100000)))
		}
		previous = current
	}
	return nil
}

func (d *nativeDeck) slideXML() (string, error) {
	var shapes, effects bytes.Buffer
	w := &nativeTiming{nextID: 4}
	for i, track := range d.tracks {
		peak, fades, err := nativeFades(track.states, d.times)
		if err != nil {
			return "", fmt.Errorf("object %d: %w", i+1, err)
		}
		shapes.WriteString(d.shapeXML(i, peak))
		shape := i + 2
		w.buf.Reset()
		if track.states[0].alpha == 0 && peak > 0 {
			w.visibility(shape, 0, false)
		}
		for _, property := range []struct {
			name  string
			value func(nativeState) float64
		}{
			{"ppt_x", func(s nativeState) float64 { return s.x / d.canvasWidth() }},
			{"ppt_y", func(s nativeState) float64 { return s.y / d.canvasHeight() }},
		} {
			values := make([]float64, len(track.states))
			for j, state := range track.states {
				values[j] = property.value(state)
			}
			w.property(shape, d.duration, property.name, values, d.times)
		}
		angles := make([]float64, len(track.states))
		for j, state := range track.states {
			angles[j] = state.rotation
		}
		w.rotations(shape, angles, d.times)
		if err := w.scales(shape, track.states, d.times); err != nil {
			return "", err
		}
		if w.buf.Len() > 0 {
			fmt.Fprintf(&effects, `<p:par><p:cTn id="%d" dur="%d" fill="hold" nodeType="withEffect"><p:stCondLst><p:cond delay="0"/></p:stCondLst><p:childTnLst>%s</p:childTnLst></p:cTn></p:par>`, w.id(), max(1, d.duration), w.buf.String())
		}
		for _, fade := range fades {
			w.buf.Reset()
			class, transition := "exit", "out"
			if fade.in {
				class, transition = "entr", "in"
				w.visibility(shape, 0, true)
			}
			duration := max(1, fade.end-fade.start)
			// Filters do not use numeric-property additive/fill behavior. In
			// particular, holding an exit filter can resurrect the opaque image.
			fmt.Fprintf(&w.buf, `<p:animEffect transition="%s" filter="fade"><p:cBhvr><p:cTn id="%d" dur="%d"/><p:tgtEl><p:spTgt spid="%d"/></p:tgtEl></p:cBhvr></p:animEffect>`, transition, w.id(), duration, shape)
			if !fade.in {
				// Match PowerPoint's last-millisecond visibility latch, without
				// extending the effect by one millisecond.
				w.visibility(shape, duration-1, false)
			}
			fmt.Fprintf(&effects, `<p:par><p:cTn id="%d" presetID="10" presetClass="%s" presetSubtype="0" dur="%d" fill="hold" nodeType="withEffect"><p:stCondLst><p:cond delay="%d"/></p:stCondLst><p:childTnLst>%s</p:childTnLst></p:cTn></p:par>`, w.id(), class, duration, fade.start, w.buf.String())
		}
	}
	var timing string
	if effects.Len() > 0 {
		timing = fmt.Sprintf(`<p:timing><p:tnLst><p:par><p:cTn id="1" dur="indefinite" restart="never" nodeType="tmRoot"><p:childTnLst><p:seq concurrent="1" nextAc="seek"><p:cTn id="2" dur="indefinite" nodeType="mainSeq"><p:childTnLst><p:par><p:cTn id="3" dur="%d" fill="hold"><p:stCondLst><p:cond delay="0"/></p:stCondLst><p:childTnLst><p:par><p:cTn id="4" dur="%d" fill="hold"><p:stCondLst><p:cond delay="0"/></p:stCondLst><p:childTnLst>%s</p:childTnLst></p:cTn></p:par></p:childTnLst></p:cTn></p:par></p:childTnLst></p:cTn><p:prevCondLst><p:cond evt="onPrev" delay="0"><p:tgtEl><p:sldTgt/></p:tgtEl></p:cond></p:prevCondLst><p:nextCondLst><p:cond evt="onNext" delay="0"><p:tgtEl><p:sldTgt/></p:tgtEl></p:cond></p:nextCondLst></p:seq></p:childTnLst></p:cTn></p:par></p:tnLst></p:timing>`, max(1, d.duration), max(1, d.duration), effects.String())
	}
	transition := fmt.Sprintf(`<p:transition advClick="0" advTm="%d"/>`, d.duration)
	if d.duration == 0 {
		transition = `<p:transition advClick="1"/>`
	}
	return fmt.Sprintf(`<p:sld %s><p:cSld name="Native animation"><p:bg><p:bgPr>%s<a:effectLst/></p:bgPr></p:bg><p:spTree>%s%s</p:spTree></p:cSld><p:clrMapOvr><a:masterClrMapping/></p:clrMapOvr>%s%s</p:sld>`, pptxNamespaces, nativeFill(d.background, 1), pptxGroup, shapes.String(), transition, timing), nil
}

func (p *pptxArchive) addNativeSlide(slide string, duration float64) error {
	if err := p.xmlPart("ppt/slides/slide1.xml", slide); err != nil {
		return err
	}
	if err := p.relations("ppt/slides/_rels/slide1.xml.rels",
		pptxRelation("rIdLayout", "slideLayout", "../slideLayouts/slideLayout1.xml"),
		pptxRelation("rIdNotes", "notesSlide", "../notesSlides/notesSlide1.xml")); err != nil {
		return err
	}
	return p.addNotes(1, pptxEscape(fmt.Sprintf("Native editable animation. Duration: %g seconds. Geometry keyframes sampled from the JS timeline; fade effects use PowerPoint interpolation. Fonts are referenced, not embedded.", duration)))
}
