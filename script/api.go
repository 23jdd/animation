package script

import (
	"errors"
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
	"github.com/23dd/animation/scene"
	"github.com/dop251/goja"
	"golang.org/x/image/font/opentype"
)

func (r *Runtime) register() error {
	var registerErr error
	set := func(name string, fn func(goja.FunctionCall) goja.Value) {
		rawName := "__animation_native_" + name
		wrapped := func(c goja.FunctionCall) goja.Value {
			r.callErr = nil
			result := fn(c)
			outcome := r.VM.NewObject()
			_ = outcome.Set("value", result)
			if err := r.takeCallError(); err != nil {
				_ = outcome.Set("error", err.Error())
			}
			return outcome
		}
		if err := r.VM.Set(rawName, wrapped); err != nil && registerErr == nil {
			registerErr = fmt.Errorf("register JavaScript function %q: %w", name, err)
			return
		}
		wrapper := fmt.Sprintf(`(function(native) {
			return function(...args) {
				const outcome = native(...args);
				if (outcome.error !== undefined) throw new TypeError(outcome.error);
				return outcome.value;
			};
		})(globalThis[%s])`, strconv.Quote(rawName))
		value, err := r.VM.RunString(wrapper)
		if err == nil {
			err = r.VM.Set(name, value)
		}
		if err != nil && registerErr == nil {
			registerErr = fmt.Errorf("register JavaScript function %q: %w", name, err)
		}
	}
	set("circle", r.jsCircle)
	set("rect", r.jsRect)
	set("rectangle", r.jsRect)
	set("line", r.jsLine)
	set("polygon", r.jsPolygon)
	set("path", r.jsPath)
	set("text", r.jsText)
	set("image", r.jsImage)
	set("group", r.jsGroup)
	set("mesh3D", r.jsMesh3D)
	set("box3D", r.jsBox3D)
	set("sphere3D", r.jsSphere3D)
	set("torus3D", r.jsTorus3D)
	set("camera3D", r.jsCamera3D)
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
	set("move3DTo", func(c goja.FunctionCall) goja.Value {
		return r.anim(animation.Move3DTo(r.mustMesh3D(c.Argument(0)), vec3(c.Argument(1))))
	})
	set("move3DBy", func(c goja.FunctionCall) goja.Value {
		return r.anim(animation.Move3DBy(r.mustMesh3D(c.Argument(0)), vec3(c.Argument(1))))
	})
	set("rotate3DTo", func(c goja.FunctionCall) goja.Value {
		return r.anim(animation.Rotate3DTo(r.mustMesh3D(c.Argument(0)), vec3(c.Argument(1))))
	})
	set("rotate3DBy", func(c goja.FunctionCall) goja.Value {
		return r.anim(animation.Rotate3DBy(r.mustMesh3D(c.Argument(0)), vec3(c.Argument(1))))
	})
	set("scale3DTo", func(c goja.FunctionCall) goja.Value {
		return r.anim(animation.Scale3DTo(r.mustMesh3D(c.Argument(0)), scaleVec3(c.Argument(1))))
	})
	set("scale3DBy", func(c goja.FunctionCall) goja.Value {
		return r.anim(animation.Scale3DBy(r.mustMesh3D(c.Argument(0)), scaleVec3(c.Argument(1))))
	})
	set("moveCamera3DTo", func(c goja.FunctionCall) goja.Value {
		return r.anim(animation.MoveCamera3DTo(r.mustCamera3D(c.Argument(0)), vec3(c.Argument(1))))
	})
	set("lookAt3D", func(c goja.FunctionCall) goja.Value {
		return r.anim(animation.LookAt3D(r.mustCamera3D(c.Argument(0)), vec3(c.Argument(1))))
	})
	set("cameraFOVTo", func(c goja.FunctionCall) goja.Value {
		return r.anim(animation.CameraFOVTo(r.mustCamera3D(c.Argument(0)), r.finiteNumber(c.Argument(1), 50, "camera FOV")))
	})
	set("fadeIn", func(c goja.FunctionCall) goja.Value { return r.anim(animation.FadeIn(r.mustObject(c.Argument(0)))) })
	set("fadeOut", func(c goja.FunctionCall) goja.Value { return r.anim(animation.FadeOut(r.mustObject(c.Argument(0)))) })
	set("opacityTo", func(c goja.FunctionCall) goja.Value {
		return r.anim(animation.OpacityTo(r.mustObject(c.Argument(0)), c.Argument(1).ToFloat()))
	})
	set("fadeTo", func(c goja.FunctionCall) goja.Value {
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
			r.fail("drawLine expects a line object")
			return goja.Undefined()
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
			r.fail("drawLineFromCenter expects a line object")
			return goja.Undefined()
		}
		return r.anim(animation.DrawLineFromCenter(target))
	})
	set("write", func(c goja.FunctionCall) goja.Value {
		target, ok := r.mustObject(c.Argument(0)).(*mobject.Text)
		if !ok {
			r.fail("write expects a text object")
			return goja.Undefined()
		}
		opa := r.finiteNumber(c.Argument(2), 1, "write opacity")
		if opa < 0 || opa > 1 {
			r.fail("write opacity must be between 0 and 1")
			return goja.Undefined()
		}
		return r.anim(animation.Write(target, cursorColor(c.Argument(1), target.Fill), opa))
	})
	set("unwrite", func(c goja.FunctionCall) goja.Value {
		target, ok := r.mustObject(c.Argument(0)).(*mobject.Text)
		if !ok {
			r.fail("unwrite expects a text object")
			return goja.Undefined()
		}
		return r.anim(animation.Unwrite(target, cursorColor(c.Argument(1), target.Fill)))
	})
	set("moveAlongPath", func(c goja.FunctionCall) goja.Value {
		p, ok := r.mustObject(c.Argument(1)).(*mobject.Path)
		if !ok {
			r.fail("moveAlongPath second argument must be a path")
			return goja.Undefined()
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
			r.fail("animateValue expects a ValueTracker")
			return goja.Undefined()
		}
		return r.anim(animation.AnimateValue(v, c.Argument(1).ToFloat()))
	})
	return registerErr
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
	if !isDefined(v) {
		r.fail("expected animation")
		return animation.New(0, animation.Linear, nil, nil, nil, nil)
	}
	a, ok := v.Export().(animation.Animation)
	if !ok {
		r.fail("expected animation")
		return animation.New(0, animation.Linear, nil, nil, nil, nil)
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
	if !isDefined(v) {
		r.fail("expected mobject")
		return mobject.NewGroup()
	}
	o := v.ToObject(r.VM).Get("__native")
	if o == nil || goja.IsUndefined(o) {
		r.fail("expected mobject")
		return mobject.NewGroup()
	}
	m, ok := o.Export().(mobject.Mobject)
	if !ok {
		r.fail("invalid mobject")
		return mobject.NewGroup()
	}
	return m
}

func (r *Runtime) mustMesh3D(v goja.Value) *mobject.Mesh3D {
	m, ok := r.mustObject(v).(*mobject.Mesh3D)
	if !ok {
		r.fail("expected a 3D mesh")
		return mobject.NewMesh3D(nil, nil)
	}
	return m
}

func (r *Runtime) mustCamera3D(v goja.Value) *scene.Camera3D {
	if !isDefined(v) {
		r.fail("expected a 3D camera")
		return &r.Scene.Camera.ThreeD
	}
	o := v.ToObject(r.VM).Get("__nativeCamera3D")
	if o == nil || goja.IsUndefined(o) {
		r.fail("expected a 3D camera")
		return &r.Scene.Camera.ThreeD
	}
	camera, ok := o.Export().(*scene.Camera3D)
	if !ok {
		r.fail("invalid 3D camera")
		return &r.Scene.Camera.ThreeD
	}
	return camera
}

func (r *Runtime) fail(message string) {
	if r.callErr == nil {
		r.callErr = errors.New(message)
	}
}

func (r *Runtime) takeCallError() error {
	err := r.callErr
	r.callErr = nil
	return err
}

func (r *Runtime) finiteNumber(v goja.Value, fallback float64, name string) float64 {
	if !isDefined(v) {
		return fallback
	}
	var n float64
	switch value := v.Export().(type) {
	case float64:
		n = value
	case int64:
		n = float64(value)
	default:
		r.fail(name + " must be a number")
		return fallback
	}
	if math.IsNaN(n) || math.IsInf(n, 0) {
		r.fail(name + " must be finite")
		return fallback
	}
	return n
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
	define("scaleX", func() float64 { return m.Base().Transform.Scale.X }, func(v float64) { m.Base().Transform.Scale.X = v })
	define("scaleY", func() float64 { return m.Base().Transform.Scale.Y }, func(v float64) { m.Base().Transform.Scale.Y = v })
	define("opacity", func() float64 { return m.Base().Opacity }, func(v float64) { m.Base().Opacity = v })
	if mesh, ok := m.(*mobject.Mesh3D); ok {
		define("x3D", func() float64 { return mesh.Transform3D.Position.X }, func(v float64) { mesh.Transform3D.Position.X = v })
		define("y3D", func() float64 { return mesh.Transform3D.Position.Y }, func(v float64) { mesh.Transform3D.Position.Y = v })
		define("z3D", func() float64 { return mesh.Transform3D.Position.Z }, func(v float64) { mesh.Transform3D.Position.Z = v })
		define("rotationX", func() float64 { return mesh.Transform3D.Rotation.X }, func(v float64) { mesh.Transform3D.Rotation.X = v })
		define("rotationY", func() float64 { return mesh.Transform3D.Rotation.Y }, func(v float64) { mesh.Transform3D.Rotation.Y = v })
		define("rotationZ", func() float64 { return mesh.Transform3D.Rotation.Z }, func(v float64) { mesh.Transform3D.Rotation.Z = v })
	}
	nativeUpdate := r.VM.ToValue(func(c goja.FunctionCall) goja.Value {
		outcome := r.VM.NewObject()
		fn, ok := goja.AssertFunction(c.Argument(0))
		if !ok {
			_ = outcome.Set("error", "update expects a function")
			return outcome
		}
		m.Base().AddUpdater(func(_ mobject.Mobject, t, dt float64) error {
			_, err := fn(goja.Undefined(), o, r.VM.ToValue(t), r.VM.ToValue(dt))
			return err
		})
		_ = outcome.Set("value", o)
		return outcome
	})
	factory, err := r.VM.RunString(`(function(native) {
		return function(callback) {
			const outcome = native(callback);
			if (outcome.error !== undefined) throw new TypeError(outcome.error);
			return outcome.value;
		};
	})`)
	if err != nil {
		r.fail("create update method: " + err.Error())
		return o
	}
	makeMethod, ok := goja.AssertFunction(factory)
	if !ok {
		r.fail("create update method: wrapper is not callable")
		return o
	}
	update, err := makeMethod(goja.Undefined(), nativeUpdate)
	if err != nil {
		r.fail("create update method: " + err.Error())
		return o
	}
	if err = o.Set("update", update); err != nil {
		r.fail("set update method: " + err.Error())
	}
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
			r.fail("font must be a .ttf or .otf file")
			return r.wrap(m)
		}
		info, err := os.Stat(fontPath)
		if err != nil {
			r.fail(fmt.Sprintf("load font %q: %v", fontValue.String(), err))
			return r.wrap(m)
		}
		if !info.Mode().IsRegular() {
			r.fail("font path must point to a file")
			return r.wrap(m)
		}
		fontData, err := os.ReadFile(fontPath)
		if err != nil {
			r.fail(fmt.Sprintf("read font %q: %v", fontValue.String(), err))
			return r.wrap(m)
		}
		if _, err = opentype.Parse(fontData); err != nil {
			r.fail(fmt.Sprintf("parse font %q: %v", fontValue.String(), err))
			return r.wrap(m)
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

func (r *Runtime) jsMesh3D(c goja.FunctionCall) goja.Value {
	o := object(c.Argument(0), r.VM)
	vertexValues := arrayValues(o.Get("vertices"), r.VM)
	vertices := make([]geometry.Vec3, 0, len(vertexValues))
	for _, value := range vertexValues {
		vertices = append(vertices, vec3(value))
	}
	faceValues := arrayValues(o.Get("faces"), r.VM)
	faces := make([][]int, 0, len(faceValues))
	for _, value := range faceValues {
		items := arrayValues(value, r.VM)
		if len(items) < 3 {
			r.fail("every 3D mesh face must contain at least three vertex indices")
			continue
		}
		face := make([]int, len(items))
		valid := true
		for i, item := range items {
			face[i] = int(item.ToInteger())
			if face[i] < 0 || face[i] >= len(vertices) {
				r.fail("3D mesh face index is outside the vertices array")
				valid = false
			}
		}
		if valid {
			faces = append(faces, face)
		}
	}
	if len(vertices) < 3 || len(faces) == 0 {
		r.fail("mesh3D requires at least three vertices and one valid face")
	}
	m := mobject.NewMesh3D(vertices, faces)
	if normalValue := o.Get("normals"); isDefined(normalValue) {
		normalValues := arrayValues(normalValue, r.VM)
		if len(normalValues) != len(vertices) {
			r.fail("mesh3D normals must contain exactly one normal for each vertex")
		} else {
			m.VertexNormals = make([]geometry.Vec3, len(normalValues))
			for i, value := range normalValues {
				m.VertexNormals[i] = vec3(value).Normalize()
			}
		}
	}
	styleMesh3D(m, o, r.VM)
	return r.wrap(m)
}

func (r *Runtime) jsBox3D(c goja.FunctionCall) goja.Value {
	o := object(c.Argument(0), r.VM)
	m := mobject.NewBox3D(propNum(o, "width", 160), propNum(o, "height", 160), propNum(o, "depth", 160))
	styleMesh3D(m, o, r.VM)
	return r.wrap(m)
}

func (r *Runtime) jsSphere3D(c goja.FunctionCall) goja.Value {
	o := object(c.Argument(0), r.VM)
	segments := int(math.Max(3, math.Min(96, propNum(o, "segments", 28))))
	rings := int(math.Max(2, math.Min(64, propNum(o, "rings", 16))))
	m := mobject.NewSphere3D(propNum(o, "radius", 120), segments, rings)
	styleMesh3D(m, o, r.VM)
	return r.wrap(m)
}

func (r *Runtime) jsTorus3D(c goja.FunctionCall) goja.Value {
	o := object(c.Argument(0), r.VM)
	segments := int(math.Max(3, math.Min(128, propNum(o, "segments", 48))))
	sides := int(math.Max(3, math.Min(64, propNum(o, "sides", 12))))
	m := mobject.NewTorus3D(propNum(o, "majorRadius", 170), propNum(o, "minorRadius", 36), segments, sides)
	styleMesh3D(m, o, r.VM)
	return r.wrap(m)
}

func (r *Runtime) jsCamera3D(c goja.FunctionCall) goja.Value {
	o := object(c.Argument(0), r.VM)
	camera := &r.Scene.Camera.ThreeD
	camera.Position = vec3Default(o.Get("position"), camera.Position)
	camera.Target = vec3Default(o.Get("target"), camera.Target)
	camera.Up = vec3Default(o.Get("up"), camera.Up).Normalize()
	camera.LightDirection = vec3Default(o.Get("lightDirection"), camera.LightDirection).Normalize()
	camera.FOV = math.Max(1, math.Min(179, propNum(o, "fov", camera.FOV)))
	camera.Near = math.Max(0.001, propNum(o, "near", camera.Near))
	camera.Ambient = math.Max(0, math.Min(1, propNum(o, "ambient", camera.Ambient)))
	camera.Diffuse = math.Max(0, math.Min(2, propNum(o, "diffuse", camera.Diffuse)))
	if v := o.Get("fogColor"); isDefined(v) {
		camera.FogColor = parseColor(v.String())
	}
	camera.FogNear = math.Max(camera.Near, propNum(o, "fogNear", camera.FogNear))
	camera.FogFar = math.Max(camera.FogNear+0.001, propNum(o, "fogFar", camera.FogFar))
	camera.FogStrength = math.Max(0, math.Min(1, propNum(o, "fogStrength", camera.FogStrength)))
	return r.wrapCamera3D(camera)
}

func (r *Runtime) wrapCamera3D(camera *scene.Camera3D) goja.Value {
	o := r.VM.NewObject()
	_ = o.Set("__nativeCamera3D", camera)
	getter := func(fn func() float64) goja.Value {
		return r.VM.ToValue(func(goja.FunctionCall) goja.Value { return r.VM.ToValue(fn()) })
	}
	setter := func(fn func(float64)) goja.Value {
		return r.VM.ToValue(func(c goja.FunctionCall) goja.Value { fn(c.Argument(0).ToFloat()); return goja.Undefined() })
	}
	define := func(name string, get func() float64, set func(float64)) {
		_ = o.DefineAccessorProperty(name, getter(get), setter(set), goja.FLAG_TRUE, goja.FLAG_TRUE)
	}
	define("x", func() float64 { return camera.Position.X }, func(v float64) { camera.Position.X = v })
	define("y", func() float64 { return camera.Position.Y }, func(v float64) { camera.Position.Y = v })
	define("z", func() float64 { return camera.Position.Z }, func(v float64) { camera.Position.Z = v })
	define("targetX", func() float64 { return camera.Target.X }, func(v float64) { camera.Target.X = v })
	define("targetY", func() float64 { return camera.Target.Y }, func(v float64) { camera.Target.Y = v })
	define("targetZ", func() float64 { return camera.Target.Z }, func(v float64) { camera.Target.Z = v })
	define("fov", func() float64 { return camera.FOV }, func(v float64) { camera.FOV = math.Max(1, math.Min(179, v)) })
	define("ambient", func() float64 { return camera.Ambient }, func(v float64) { camera.Ambient = math.Max(0, math.Min(1, v)) })
	define("diffuse", func() float64 { return camera.Diffuse }, func(v float64) { camera.Diffuse = math.Max(0, math.Min(2, v)) })
	define("fogNear", func() float64 { return camera.FogNear }, func(v float64) { camera.FogNear = math.Max(camera.Near, v) })
	define("fogFar", func() float64 { return camera.FogFar }, func(v float64) { camera.FogFar = math.Max(camera.FogNear+0.001, v) })
	define("fogStrength", func() float64 { return camera.FogStrength }, func(v float64) { camera.FogStrength = math.Max(0, math.Min(1, v)) })
	return o
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
func vec3(v goja.Value) geometry.Vec3                        { return vec3Default(v, geometry.Vec3{}) }
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
func scaleVec3(v goja.Value) geometry.Vec3 {
	if v == nil || goja.IsUndefined(v) || goja.IsNull(v) {
		return geometry.V3(1, 1, 1)
	}
	if _, ok := v.Export().(float64); ok {
		x := v.ToFloat()
		return geometry.V3(x, x, x)
	}
	if _, ok := v.Export().(int64); ok {
		x := v.ToFloat()
		return geometry.V3(x, x, x)
	}
	return vec3Default(v, geometry.V3(1, 1, 1))
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
func vec3Default(v goja.Value, d geometry.Vec3) geometry.Vec3 {
	if v == nil || goja.IsUndefined(v) || goja.IsNull(v) {
		return d
	}
	o := v.ToObject(nil)
	if o == nil {
		return d
	}
	return geometry.V3(number(o.Get("0"), d.X), number(o.Get("1"), d.Y), number(o.Get("2"), d.Z))
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

func styleMesh3D(m *mobject.Mesh3D, o *goja.Object, vm *goja.Runtime) {
	b := m.Base()
	m.Transform3D.Position = vec3Default(o.Get("position"), m.Transform3D.Position)
	m.Transform3D.Rotation = vec3Default(o.Get("rotation"), m.Transform3D.Rotation)
	m.Transform3D.Scale = scaleVec3(o.Get("scale"))
	b.Opacity = propNum(o, "opacity", 1)
	b.StrokeWidth = propNum(o, "strokeWidth", 1)
	if v := o.Get("fill"); isDefined(v) {
		b.Fill = parseColor(v.String())
	}
	if v := o.Get("stroke"); isDefined(v) {
		b.Stroke = parseColor(v.String())
	}
	m.Wireframe = propBool(o, "wireframe", m.Wireframe)
	m.CullBackfaces = propBool(o, "cullBackfaces", m.CullBackfaces)
	m.Shading = propBool(o, "shading", m.Shading)
	m.SmoothShading = propBool(o, "smoothShading", m.SmoothShading)
	m.Emissive = math.Max(0, math.Min(1, propNum(o, "emissive", m.Emissive)))
	m.Specular = math.Max(0, math.Min(2, propNum(o, "specular", m.Specular)))
	m.Shininess = math.Max(1, math.Min(256, propNum(o, "shininess", m.Shininess)))
	m.GlowRadius = math.Max(0, propNum(o, "glow", 0))
	if v := o.Get("glowColor"); isDefined(v) {
		m.GlowColor = parseColor(v.String())
	}
	for _, value := range arrayValues(o.Get("faceColors"), vm) {
		m.FaceColors = append(m.FaceColors, parseColor(value.String()))
	}
}

func propBool(o *goja.Object, name string, fallback bool) bool {
	v := o.Get(name)
	if !isDefined(v) {
		return fallback
	}
	return v.ToBoolean()
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
