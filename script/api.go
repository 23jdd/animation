package script

import (
	"fmt"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/23dd/animation/animation"
	"github.com/23dd/animation/geometry"
	"github.com/23dd/animation/mobject"
	"github.com/dop251/goja"
	"golang.org/x/image/font/opentype"
)

func (r *Runtime) register() {
	set := func(name string, fn func(goja.FunctionCall) goja.Value) {
		if err := r.VM.Set(name, fn); err != nil {
			panic(err)
		}
	}
	set("circle", r.jsCircle)
	set("rect", r.jsRect)
	set("rectangle", r.jsRect)
	set("line", r.jsLine)
	set("polygon", r.jsPolygon)
	set("path", r.jsPath)
	set("text", r.jsText)
	set("group", r.jsGroup)
	set("add", func(c goja.FunctionCall) goja.Value {
		for _, v := range c.Arguments {
			r.Scene.Add(r.mustObject(v))
		}
		return goja.Undefined()
	})
	set("remove", func(c goja.FunctionCall) goja.Value {
		for _, v := range c.Arguments {
			r.Scene.Remove(r.mustObject(v))
		}
		return goja.Undefined()
	})
	set("play", r.jsPlay)
	set("wait", func(c goja.FunctionCall) goja.Value {
		d := number(c.Argument(0), 1)
		r.cursor += math.Max(0, d)
		r.Timeline.Wait(math.Max(0, d))
		return goja.Undefined()
	})
	set("moveTo", func(c goja.FunctionCall) goja.Value {
		return r.anim(animation.MoveTo(r.mustObject(c.Argument(0)), vec(c.Argument(1))))
	})
	set("moveBy", func(c goja.FunctionCall) goja.Value {
		return r.anim(animation.MoveBy(r.mustObject(c.Argument(0)), vec(c.Argument(1))))
	})
	set("scaleTo", func(c goja.FunctionCall) goja.Value {
		return r.anim(animation.ScaleTo(r.mustObject(c.Argument(0)), scaleVec(c.Argument(1))))
	})
	set("scaleBy", func(c goja.FunctionCall) goja.Value {
		return r.anim(animation.ScaleBy(r.mustObject(c.Argument(0)), scaleVec(c.Argument(1))))
	})
	set("rotateTo", func(c goja.FunctionCall) goja.Value {
		return r.anim(animation.RotateTo(r.mustObject(c.Argument(0)), c.Argument(1).ToFloat()))
	})
	set("rotateBy", func(c goja.FunctionCall) goja.Value {
		return r.anim(animation.RotateBy(r.mustObject(c.Argument(0)), c.Argument(1).ToFloat()))
	})
	set("fadeIn", func(c goja.FunctionCall) goja.Value { return r.anim(animation.FadeIn(r.mustObject(c.Argument(0)))) })
	set("fadeOut", func(c goja.FunctionCall) goja.Value { return r.anim(animation.FadeOut(r.mustObject(c.Argument(0)))) })
	set("opacityTo", func(c goja.FunctionCall) goja.Value {
		return r.anim(animation.OpacityTo(r.mustObject(c.Argument(0)), c.Argument(1).ToFloat()))
	})
	set("colorTo", func(c goja.FunctionCall) goja.Value {
		return r.anim(animation.ColorTo(r.mustObject(c.Argument(0)), parseColor(c.Argument(1).String())))
	})
	set("create", func(c goja.FunctionCall) goja.Value { return r.anim(animation.Create(r.mustObject(c.Argument(0)))) })
	set("uncreate", func(c goja.FunctionCall) goja.Value { return r.anim(animation.Uncreate(r.mustObject(c.Argument(0)))) })
	set("drawLine", func(c goja.FunctionCall) goja.Value {
		target, ok := r.mustObject(c.Argument(0)).(*mobject.Line)
		if !ok {
			panic(r.VM.NewTypeError("drawLine expects a line object"))
		}
		var tip mobject.Mobject
		if v := c.Argument(1); isDefined(v) {
			tip = r.mustObject(v)
		}
		return r.anim(animation.DrawLine(target, tip))
	})
	set("drawLineFromCenter", func(c goja.FunctionCall) goja.Value {
		target, ok := r.mustObject(c.Argument(0)).(*mobject.Line)
		if !ok {
			panic(r.VM.NewTypeError("drawLineFromCenter expects a line object"))
		}
		return r.anim(animation.DrawLineFromCenter(target))
	})
	set("write", func(c goja.FunctionCall) goja.Value {
		target, ok := r.mustObject(c.Argument(0)).(*mobject.Text)
		if !ok {
			panic(r.VM.NewTypeError("write expects a text object"))
		}
		return r.anim(animation.Write(target, cursorColor(c.Argument(1), target.Fill)))
	})
	set("unwrite", func(c goja.FunctionCall) goja.Value {
		target, ok := r.mustObject(c.Argument(0)).(*mobject.Text)
		if !ok {
			panic(r.VM.NewTypeError("unwrite expects a text object"))
		}
		return r.anim(animation.Unwrite(target, cursorColor(c.Argument(1), target.Fill)))
	})
	set("moveAlongPath", func(c goja.FunctionCall) goja.Value {
		p, ok := r.mustObject(c.Argument(1)).(*mobject.Path)
		if !ok {
			panic(r.VM.NewTypeError("second argument must be a path"))
		}
		return r.anim(animation.MoveAlongPath(r.mustObject(c.Argument(0)), p.Geometry))
	})
	set("transform", func(c goja.FunctionCall) goja.Value {
		return r.anim(animation.Transform(r.mustObject(c.Argument(0)), r.mustObject(c.Argument(1))))
	})
	set("replacementTransform", func(c goja.FunctionCall) goja.Value {
		return r.anim(animation.ReplacementTransform(r.mustObject(c.Argument(0)), r.mustObject(c.Argument(1))))
	})
	set("growFromCenter", func(c goja.FunctionCall) goja.Value {
		return r.anim(animation.GrowFromCenter(r.mustObject(c.Argument(0))))
	})
	set("shrinkToCenter", func(c goja.FunctionCall) goja.Value {
		return r.anim(animation.ShrinkToCenter(r.mustObject(c.Argument(0))))
	})
	set("growFromPoint", func(c goja.FunctionCall) goja.Value {
		return r.anim(animation.GrowFromPoint(r.mustObject(c.Argument(0)), vecDefault(c.Argument(1), geometry.V2(0, 0))))
	})
	set("growFromEdge", func(c goja.FunctionCall) goja.Value {
		return r.anim(animation.GrowFromEdge(r.mustObject(c.Argument(0)), stringValue(c.Argument(1), "up"), number(c.Argument(2), 150)))
	})
	set("spinIn", func(c goja.FunctionCall) goja.Value {
		return r.anim(animation.SpinIn(r.mustObject(c.Argument(0))))
	})
	set("spinOut", func(c goja.FunctionCall) goja.Value {
		return r.anim(animation.SpinOut(r.mustObject(c.Argument(0))))
	})
	set("flash", func(c goja.FunctionCall) goja.Value {
		highlight := color.RGBA{255, 255, 255, 255}
		if v := c.Argument(1); isDefined(v) {
			highlight = parseColor(v.String())
		}
		return r.anim(animation.Flash(r.mustObject(c.Argument(0)), highlight))
	})
	set("indicate", func(c goja.FunctionCall) goja.Value {
		scale := number(c.Argument(1), 1.3)
		highlight := color.RGBA{255, 255, 255, 255}
		if v := c.Argument(2); isDefined(v) {
			highlight = parseColor(v.String())
		}
		return r.anim(animation.Indicate(r.mustObject(c.Argument(0)), scale, highlight))
	})
	set("blink", func(c goja.FunctionCall) goja.Value {
		return r.anim(animation.Blink(r.mustObject(c.Argument(0)), number(c.Argument(1), 1)))
	})
	set("wiggle", func(c goja.FunctionCall) goja.Value {
		return r.anim(animation.Wiggle(r.mustObject(c.Argument(0)), number(c.Argument(1), 0.3), number(c.Argument(2), 2)))
	})
	set("wave", func(c goja.FunctionCall) goja.Value {
		return r.anim(animation.Wave(r.mustObject(c.Argument(0)), number(c.Argument(1), 20)))
	})
	set("zoomIn", func(c goja.FunctionCall) goja.Value {
		return r.anim(animation.ZoomIn(r.mustObject(c.Argument(0)), number(c.Argument(1), 2.5)))
	})
	set("zoomOut", func(c goja.FunctionCall) goja.Value {
		return r.anim(animation.ZoomOut(r.mustObject(c.Argument(0)), number(c.Argument(1), 2.5)))
	})
	set("focusOn", func(c goja.FunctionCall) goja.Value {
		return r.anim(animation.FocusOn(r.mustObject(c.Argument(0)), number(c.Argument(1), 1.25)))
	})
	set("sequence", func(c goja.FunctionCall) goja.Value { return r.anim(animation.Sequence(r.animations(c.Arguments)...)) })
	set("parallel", func(c goja.FunctionCall) goja.Value { return r.anim(animation.Parallel(r.animations(c.Arguments)...)) })
	set("stagger", func(c goja.FunctionCall) goja.Value {
		return r.anim(animation.Stagger(c.Argument(0).ToFloat(), r.animations(c.Arguments[1:])...))
	})
	set("repeat", func(c goja.FunctionCall) goja.Value {
		return r.anim(animation.Repeat(r.mustAnimation(c.Argument(0)), int(c.Argument(1).ToInteger())))
	})
	set("yoyo", func(c goja.FunctionCall) goja.Value {
		return r.anim(animation.Yoyo(r.mustAnimation(c.Argument(0)), int(c.Argument(1).ToInteger())))
	})
	set("loop", func(c goja.FunctionCall) goja.Value {
		return r.anim(animation.Repeat(r.mustAnimation(c.Argument(0)), int(number(c.Argument(1), 2))))
	})
	set("value", func(c goja.FunctionCall) goja.Value {
		return r.wrapValue(animation.NewValueTracker(number(c.Argument(0), 0)))
	})
	set("animateValue", func(c goja.FunctionCall) goja.Value {
		native := c.Argument(0).ToObject(r.VM).Get("__native")
		v, ok := native.Export().(*animation.ValueTracker)
		if !ok {
			panic(r.VM.NewTypeError("expected ValueTracker"))
		}
		return r.anim(animation.AnimateValue(v, c.Argument(1).ToFloat()))
	})
}

func (r *Runtime) wrapValue(v *animation.ValueTracker) goja.Value {
	o := r.VM.NewObject()
	_ = o.Set("__native", v)
	get := r.VM.ToValue(func(goja.FunctionCall) goja.Value { return r.VM.ToValue(v.Value) })
	set := r.VM.ToValue(func(c goja.FunctionCall) goja.Value { v.Value = c.Argument(0).ToFloat(); return goja.Undefined() })
	_ = o.DefineAccessorProperty("value", get, set, goja.FLAG_TRUE, goja.FLAG_TRUE)
	return o
}

func (r *Runtime) anim(a animation.Animation) goja.Value { return r.VM.ToValue(a) }
func (r *Runtime) mustAnimation(v goja.Value) animation.Animation {
	a, ok := v.Export().(animation.Animation)
	if !ok {
		panic(r.VM.NewTypeError("expected animation"))
	}
	return a
}
func (r *Runtime) animations(values []goja.Value) []animation.Animation {
	out := make([]animation.Animation, 0, len(values))
	for _, v := range values {
		out = append(out, r.mustAnimation(v))
	}
	return out
}
func (r *Runtime) mustObject(v goja.Value) mobject.Mobject {
	o := v.ToObject(r.VM).Get("__native")
	if o == nil || goja.IsUndefined(o) {
		panic(r.VM.NewTypeError("expected mobject"))
	}
	m, ok := o.Export().(mobject.Mobject)
	if !ok {
		panic(r.VM.NewTypeError("invalid mobject"))
	}
	return m
}
func (r *Runtime) jsPlay(c goja.FunctionCall) goja.Value {
	args := c.Arguments
	duration := -1.0
	ease := animation.Linear
	hasOptions := false
	if len(args) > 0 {
		last := args[len(args)-1]
		_, isAnimation := last.Export().(animation.Animation)
		obj := last.ToObject(r.VM)
		var durationValue, easeValue goja.Value
		if obj != nil {
			durationValue, easeValue = obj.Get("duration"), obj.Get("ease")
		}
		if !isAnimation && (isDefined(durationValue) || isDefined(easeValue)) {
			hasOptions = true
			duration = number(durationValue, -1)
			ease = animation.EasingByName(stringValue(easeValue, "linear"))
			args = args[:len(args)-1]
		}
	}
	a := animation.Parallel(r.animations(args)...)
	if len(args) == 1 {
		a = r.mustAnimation(args[0])
	}
	if hasOptions {
		if duration < 0 {
			duration = a.Duration()
		}
		a = animation.WithTiming(a, duration, ease)
	}
	r.Timeline.Add(r.cursor, a)
	r.cursor += a.Duration()
	return goja.Undefined()
}

func (r *Runtime) wrap(m mobject.Mobject) goja.Value {
	o := r.VM.NewObject()
	_ = o.Set("__native", m)
	getter := func(fn func() float64) goja.Value {
		return r.VM.ToValue(func(goja.FunctionCall) goja.Value { return r.VM.ToValue(fn()) })
	}
	setter := func(fn func(float64)) goja.Value {
		return r.VM.ToValue(func(c goja.FunctionCall) goja.Value { fn(c.Argument(0).ToFloat()); return goja.Undefined() })
	}
	define := func(name string, get func() float64, set func(float64)) {
		_ = o.DefineAccessorProperty(name, getter(get), setter(set), goja.FLAG_TRUE, goja.FLAG_TRUE)
	}
	define("x", func() float64 { return m.Base().Transform.Position.X }, func(v float64) { m.Base().Transform.Position.X = v })
	define("y", func() float64 { return m.Base().Transform.Position.Y }, func(v float64) { m.Base().Transform.Position.Y = v })
	define("rotation", func() float64 { return m.Base().Transform.Rotation }, func(v float64) { m.Base().Transform.Rotation = v })
	define("opacity", func() float64 { return m.Base().Opacity }, func(v float64) { m.Base().Opacity = v })
	_ = o.Set("update", func(c goja.FunctionCall) goja.Value {
		fn, ok := goja.AssertFunction(c.Argument(0))
		if !ok {
			panic(r.VM.NewTypeError("update expects a function"))
		}
		m.Base().AddUpdater(func(_ mobject.Mobject, t, dt float64) {
			if _, err := fn(goja.Undefined(), o, r.VM.ToValue(t), r.VM.ToValue(dt)); err != nil {
				panic(err)
			}
		})
		return o
	})
	return o
}
func (r *Runtime) jsCircle(c goja.FunctionCall) goja.Value {
	o := object(c.Argument(0), r.VM)
	m := mobject.NewCircle(propNum(o, "radius", 50))
	style(m, o)
	return r.wrap(m)
}
func (r *Runtime) jsRect(c goja.FunctionCall) goja.Value {
	o := object(c.Argument(0), r.VM)
	m := mobject.NewRectangle(propNum(o, "width", 100), propNum(o, "height", 100))
	m.CornerRadius = propNum(o, "radius", 0)
	style(m, o)
	return r.wrap(m)
}
func (r *Runtime) jsLine(c goja.FunctionCall) goja.Value {
	o := object(c.Argument(0), r.VM)
	m := mobject.NewLine(vecDefault(o.Get("start"), geometry.V2(-50, 0)), vecDefault(o.Get("end"), geometry.V2(50, 0)))
	style(m, o)
	return r.wrap(m)
}
func (r *Runtime) jsPolygon(c goja.FunctionCall) goja.Value {
	o := object(c.Argument(0), r.VM)
	var pts []geometry.Vec2
	for _, v := range arrayValues(o.Get("points"), r.VM) {
		pts = append(pts, vec(v))
	}
	m := mobject.NewPolygon(pts...)
	style(m, o)
	return r.wrap(m)
}
func (r *Runtime) jsPath(c goja.FunctionCall) goja.Value {
	o := object(c.Argument(0), r.VM)
	var p geometry.Path
	for _, v := range arrayValues(o.Get("commands"), r.VM) {
		a := arrayValues(v, r.VM)
		if len(a) == 0 {
			continue
		}
		switch strings.ToUpper(a[0].String()) {
		case "M":
			p.MoveTo(geometry.V2(a[1].ToFloat(), a[2].ToFloat()))
		case "L":
			p.LineTo(geometry.V2(a[1].ToFloat(), a[2].ToFloat()))
		case "Q":
			p.QuadraticTo(geometry.V2(a[1].ToFloat(), a[2].ToFloat()), geometry.V2(a[3].ToFloat(), a[4].ToFloat()))
		case "C":
			p.CubicTo(geometry.V2(a[1].ToFloat(), a[2].ToFloat()), geometry.V2(a[3].ToFloat(), a[4].ToFloat()), geometry.V2(a[5].ToFloat(), a[6].ToFloat()))
		case "Z":
			p.Close()
		}
	}
	m := mobject.NewPath(p)
	style(m, o)
	return r.wrap(m)
}
func (r *Runtime) jsText(c goja.FunctionCall) goja.Value {
	o := object(c.Argument(1), r.VM)
	m := mobject.NewText(c.Argument(0).String(), propNum(o, "size", 24))
	if fontValue := o.Get("font"); isDefined(fontValue) {
		fontPath := fontValue.String()
		if !filepath.IsAbs(fontPath) {
			fontPath = filepath.Join(r.BaseDir, fontPath)
		}
		fontPath = filepath.Clean(fontPath)
		ext := strings.ToLower(filepath.Ext(fontPath))
		if ext != ".ttf" && ext != ".otf" {
			panic(r.VM.NewTypeError("font must be a .ttf or .otf file"))
		}
		info, err := os.Stat(fontPath)
		if err != nil {
			panic(r.VM.NewGoError(fmt.Errorf("load font %q: %w", fontValue.String(), err)))
		}
		if !info.Mode().IsRegular() {
			panic(r.VM.NewTypeError("font path must point to a file"))
		}
		fontData, err := os.ReadFile(fontPath)
		if err != nil {
			panic(r.VM.NewGoError(fmt.Errorf("read font %q: %w", fontValue.String(), err)))
		}
		if _, err = opentype.Parse(fontData); err != nil {
			panic(r.VM.NewGoError(fmt.Errorf("parse font %q: %w", fontValue.String(), err)))
		}
		m.Font = fontPath
		m.FontVersion = info.ModTime().UnixNano() ^ info.Size()
	}
	style(m, o)
	return r.wrap(m)
}
func (r *Runtime) jsGroup(c goja.FunctionCall) goja.Value {
	return r.wrap(mobject.NewGroup(func() []mobject.Mobject {
		var x []mobject.Mobject
		for _, v := range c.Arguments {
			x = append(x, r.mustObject(v))
		}
		return x
	}()...))
}

func object(v goja.Value, vm *goja.Runtime) *goja.Object {
	if v == nil || goja.IsUndefined(v) || goja.IsNull(v) {
		return vm.NewObject()
	}
	return v.ToObject(vm)
}
func number(v goja.Value, d float64) float64 {
	if v == nil || goja.IsUndefined(v) || goja.IsNull(v) {
		return d
	}
	return v.ToFloat()
}
func stringValue(v goja.Value, d string) string {
	if v == nil || goja.IsUndefined(v) || goja.IsNull(v) {
		return d
	}
	return v.String()
}
func isDefined(v goja.Value) bool                            { return v != nil && !goja.IsUndefined(v) && !goja.IsNull(v) }
func propNum(o *goja.Object, name string, d float64) float64 { return number(o.Get(name), d) }
func vec(v goja.Value) geometry.Vec2                         { return vecDefault(v, geometry.Vec2{}) }
func scaleVec(v goja.Value) geometry.Vec2 {
	if v == nil || goja.IsUndefined(v) || goja.IsNull(v) {
		return geometry.V2(1, 1)
	}
	if _, ok := v.Export().(float64); ok {
		x := v.ToFloat()
		return geometry.V2(x, x)
	}
	if _, ok := v.Export().(int64); ok {
		x := v.ToFloat()
		return geometry.V2(x, x)
	}
	return vecDefault(v, geometry.V2(1, 1))
}
func vecDefault(v goja.Value, d geometry.Vec2) geometry.Vec2 {
	if v == nil || goja.IsUndefined(v) || goja.IsNull(v) {
		return d
	}
	o := v.ToObject(nil)
	if o == nil {
		return d
	}
	return geometry.V2(number(o.Get("0"), d.X), number(o.Get("1"), d.Y))
}
func arrayValues(v goja.Value, vm *goja.Runtime) []goja.Value {
	if v == nil || goja.IsUndefined(v) || goja.IsNull(v) {
		return nil
	}
	o := v.ToObject(vm)
	n := int(o.Get("length").ToInteger())
	out := make([]goja.Value, n)
	for i := range out {
		out[i] = o.Get(strconv.Itoa(i))
	}
	return out
}
func style(m mobject.Mobject, o *goja.Object) {
	b := m.Base()
	b.Transform.Position = vecDefault(o.Get("position"), b.Transform.Position)
	b.Transform.Rotation = propNum(o, "rotation", 0)
	b.Transform.Scale = scaleVec(o.Get("scale"))
	b.Opacity = propNum(o, "opacity", 1)
	b.StrokeWidth = propNum(o, "strokeWidth", 2)
	if v := o.Get("fill"); v != nil && !goja.IsUndefined(v) {
		b.Fill = parseColor(v.String())
	}
	if v := o.Get("stroke"); v != nil && !goja.IsUndefined(v) {
		b.Stroke = parseColor(v.String())
	}
}
func parseColor(s string) color.RGBA {
	s = strings.TrimSpace(strings.TrimPrefix(s, "#"))
	if len(s) == 3 {
		s = fmt.Sprintf("%c%c%c%c%c%c", s[0], s[0], s[1], s[1], s[2], s[2])
	}
	if len(s) == 6 {
		s += "ff"
	}
	v, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return color.RGBA{255, 255, 255, 255}
	}
	return color.RGBA{uint8(v >> 24), uint8(v >> 16), uint8(v >> 8), uint8(v)}
}

// cursorColor resolves the caret color for write/unwrite. Undefined uses the
// fallback (the text fill), null disables the caret, and any other value is
// parsed as a color.
func cursorColor(v goja.Value, fallback color.RGBA) color.RGBA {
	if v == nil || goja.IsUndefined(v) {
		return fallback
	}
	if goja.IsNull(v) {
		return color.RGBA{}
	}
	return parseColor(v.String())
}
